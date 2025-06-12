package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/xhd2015/kool/tools/ai"
	"github.com/xhd2015/kool/tools/create"
	"github.com/xhd2015/kool/tools/dlv"
	"github.com/xhd2015/kool/tools/git"
	go_tools "github.com/xhd2015/kool/tools/go"
	"github.com/xhd2015/kool/tools/go/run"
	"github.com/xhd2015/kool/tools/go/with_go"
	"github.com/xhd2015/kool/tools/http"
	"github.com/xhd2015/kool/tools/port"
	"github.com/xhd2015/kool/tools/react"
	"github.com/xhd2015/kool/tools/rules"
	xgo_cmd "github.com/xhd2015/xgo/support/cmd"
	"golang.org/x/term"
)

// install: go build -o $GOPATH/bin/kool
const help = `
kool facalitate the use of common CLI tools.

Usage: kool <cmd> [OPTIONS]          execute command
       kool ? <question>             search for the question in it's knowledge

Commands category:
  go
  git
  vscode
  http
  ai

Utility commands:
  kill-port <port>                   kill process on the given port
  check-port-ready <port>            check if the port is ready
  help                               show help message

String commands:
  unquote                            unquote string
  compress                           compress json string
  lines
    uniq                             uniq lines without sorting, last preserved 
	reverse                          reverse lines
	sort                             sort lines
  NOTE: lines accept multiple commands toggether: kool lines uniq sort

VSCode:
  vscode                             print example vscode configs
  vscode debug-go <prog> [args...]   print vscode config for debugging go program with args

Project:  
  create <template> <project-name>   create new project
  snippet <name>                     print snippet
  go
    replace <dir>                    replace go module in the given directory
    update <dir>                     update to the latest tag of the module in dir
    inspect <pkg> <T>                inspect the given package and type
	run --debug <flags> [args...]    run the given program with debug mode
    example
	  parse-flag                     code snippet for parsing flag
  git
    tag-next                         tag next
	show-tag [<dir>]                 show the tag of the given directory
	show-exclude                     show the exclude rules
	show-children <commit>           show the children of the given commit
	check-merge <ref1> <ref2> ...    check if refs are merged into HEAD
  http
    serve [--port <port>] [DIR]      start a static HTTP server (default port: 8080)
                                     DIR is the directory to serve (default: current directory)
  with
    goX.Y <commands>                install goX.Y and execute the given command
  with-goroot <GOROOT> <commands>   set GOROOT and execute the given command
  rule,rules
    add <file>                       add a rule file to ~/.kool/rules/files/
    list                             list all available rule files
    use <file>                       copy a rule file to .cursor/rules/ if not exists
    dir                              show the rules directory location
    rm <file>                        remove a rule file from ~/.kool/rules/files/

Help:
  kool ?
  kool ? mermaid to image
  kool help                        show help message
  kool <cmd> --help                show help message for the given command
`

type ExitCodeAware interface {
	SilenceExitCode() int
}

//go:embed pkgs/flag/parse.go
var flagSnippet string

// install: go build -o `which kool` ./
func main() {
	err := handle(os.Args[1:])
	if err != nil {
		// Check for custom exit code
		var exitCodeErr ExitCodeAware
		if errors.As(err, &exitCodeErr) {
			os.Exit(exitCodeErr.SilenceExitCode())
			return
		}

		// Regular error
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	var arg0 string
	if len(args) > 0 {
		arg0 = args[0]
	}

	// TTY:     go run ./
	// NOT TTY: echo yes | go run ./
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))

	var cmd string
	switch arg0 {
	case "help", "--help":
		fmt.Println(strings.TrimSpace(strings.ReplaceAll(help, "\t", "    ")))
		return nil
	case "upgrade":
		return xgo_cmd.Debug().Run("go", "install", "github.com/xhd2015/kool@latest")
	case "sample":
		cmd = arg0
		args = args[1:]
	case "vscode":
		return handleVscode(args[1:])
	case "unquote":
		args = args[1:]
		var str string
		if len(args) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			str = string(data)
		} else {
			str = strings.Join(args, " ")
		}
		unquoteStr, err := strconv.Unquote(str)
		if err != nil {
			return err
		}
		fmt.Println(unquoteStr)
		return nil
	case "compress":
		cmd = arg0
		args = args[1:]
	case "create":
		return create.Handle(args[1:])
	case "snippet":
		return handleSnippet(args[1:])
	case "go":
		return go_tools.Handle(args[1:], flagSnippet)
	case "go-replace":
		return go_tools.HandleReplace(args[1:])
	case "go-update":
		return go_tools.HandleUpdate(args[1:])
	case "go-resolve":
		return go_tools.HandleResolve(args[1:])
	case "dlv":
		return dlv.Handle(args[1:])
	case "git":
		return git.Handle(args[1:])
	case "http":
		return http.Handle(args[1:])
	case "with":
		return handleWith(args[1:])
	case "with-go":
		return handleWithGo(args[1:])
	case "with-goroot":
		return handleWithGoroot(args[1:])
	case "ai":
		return ai.Handle(args[1:])
	case "lines":
		return handleLines(args[1:])
	case "rule", "rules":
		return rules.Handle(args[1:])
	case "check-port-ready":
		return port.CheckReady(args[1:])
	case "react":
		return react.Handle(args[1:])
	case "kill-port":
		// lsof -iTCP:15000 -sTCP:LISTEN -t
		//   -iTCP:15000: only TCP listen on port 15000
		//   -sTCP:LISTEN: only show listening socket
		//   -t: only show pid
		killPortArgs := args[1:]
		if len(killPortArgs) == 0 {
			return fmt.Errorf("usage: kool kill-port <port>")
		}
		port, err := strconv.Atoi(killPortArgs[0])
		if err != nil {
			return err
		}
		pidOutput, err := exec.Command("lsof", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-t").Output()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
				fmt.Fprintf(os.Stderr, "no process on port %d\n", port)
				return nil
			}
			return err
		}
		pid := strings.TrimSpace(string(pidOutput))
		if pid == "" {
			fmt.Fprintf(os.Stderr, "no process on port %d\n", port)
			return nil
		}
		fmt.Printf("kill -9 %s\n", pid)
		killCmd := exec.Command("kill", "-9", pid)
		killCmd.Stdout = os.Stdout
		killCmd.Stderr = os.Stderr
		return killCmd.Run()
	case "?":
		return handleQuestion(args[1:])
	default:
		if strings.HasPrefix(arg0, "with-") {
			withCmd := strings.TrimPrefix(arg0, "with-")
			if withCmd == "" {
				return fmt.Errorf("example: kool with-go1.23")
			}
			return handleWithCmd(withCmd, args[1:])
		}

		// capture unknown command
		if arg0 != "" {
			return fmt.Errorf("unknown command: %s", arg0)
		}
	}

	var remainArgs []string
	n := len(args)
	for i := 0; i < n; i++ {
		if args[i] == "--help" {
			fmt.Println(strings.TrimSpace(help))
			return nil
		}
		if args[i] == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if strings.HasPrefix(args[i], "-") {
			return fmt.Errorf("unrecognized flag: %v", args[i])
		}
		remainArgs = append(remainArgs, args[i])
	}

	if isTTY && n == 0 {
		fmt.Println(strings.TrimSpace(help))
		return nil
	}

	if !isTTY {
		if cmd == "compress" {
			var v interface{}
			decoder := json.NewDecoder(os.Stdin)
			decoder.UseNumber()
			err := decoder.Decode(&v)
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetEscapeHTML(false)
			err = encoder.Encode(v)
			if err != nil {
				return err
			}
			return nil
		}
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		if enclosedBy(data, [][2]byte{{'"', '"'}}) {
			if unquoted, err := strconv.Unquote(string(data)); err == nil {
				fmt.Println(unquoted)
				return nil
			}
		}
		if enclosedBy(data, [][2]byte{{'{', '}'}, {'[', ']'}}) {
			// json
			if cmd == "sample" {
				var match string
				if len(remainArgs) > 0 {
					match = remainArgs[0]
				}
				return sampleJSON(data, match)
			}
			// try pretty
			if v, err := decodeJSON(data); err == nil {
				if data, err := prettyJSON(v); err == nil {
					fmt.Println(string(data))
					return nil
				}
			}
			return nil
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fmt.Println(strcase.ToSnake(line))
			fmt.Println(strcase.ToCamel(line))
			fmt.Println(strcase.ToLowerCamel(line))
			fmt.Println(strcase.ToScreamingSnake(line))
			fmt.Println(strcase.ToKebab(line))
		}
	}

	return nil
}

func sampleJSON(data []byte, match string) error {
	v, err := decodeJSON(data)
	if err != nil {
		return err
	}

	_, sample := traverseSample(v, match)
	sampleData, err := prettyJSON(sample)
	if err != nil {
		return err
	}
	fmt.Println(string(sampleData))
	return nil
}
func prettyJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func traverseSample(v interface{}, match string) (bool, interface{}) {
	if v == nil {
		return match == "", nil
	}
	switch v := v.(type) {
	case []interface{}:
		var newV []interface{}
		var hasMatch bool
		for _, e := range v {
			ok, x := traverseSample(e, match)
			if !ok {
				continue
			}
			hasMatch = true
			newV = append(newV, x)
			if match == "" && len(newV) >= 2 {
				break
			}
		}
		return hasMatch, newV
	case map[string]interface{}:
		var hasAnyMatch bool
		newMap := make(map[string]interface{}, len(v))
		for k, e := range v {
			ok, x := traverseSample(e, match)
			if ok {
				hasAnyMatch = true
			}
			newMap[k] = x
		}
		return hasAnyMatch, newMap
	case string:
		hasMatch := match == "" || strings.Contains(v, match)
		return hasMatch, v
	default:
		return match == "", v
	}
}

func decodeJSON(data []byte) (interface{}, error) {
	var v interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func enclosedBy(data []byte, pairs [][2]byte) bool {
	if len(data) < 2 {
		return false
	}
	i := 0
	n := len(data)
	for ; i < n && isSpace(data[i]); i++ {
	}
	if i >= n {
		return false
	}
	var match [2]byte
	var found bool
	for _, pair := range pairs {
		if data[i] == pair[0] {
			match = pair
			found = true
			break
		}
	}
	if !found {
		return false
	}
	j := n - 1
	for ; j > i && isSpace(data[j]); j-- {
	}
	if j <= i {
		return false
	}
	return data[j] == match[1]
}
func isSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

func handleWith(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("example: kool with go1.23")
	}
	return handleWithCmd(args[0], args[1:])
}

func handleWithCmd(cmd string, args []string) error {
	if strings.HasPrefix(cmd, "go") {
		// TODO: use current go if match
		goroot, err := resolveGoroot(cmd)
		if err != nil {
			return err
		}
		return execGoroot(goroot, args)
	}
	return fmt.Errorf("unknown command: %s", cmd)
}

func handleWithGo(args []string) error {
	if len(args) == 0 {
		return errors.New("example: kool with-go [GOROOT=<X> | goX.Y] ...")
	}
	var goroot string
	var err error
	arg0 := args[0]
	if arg0 == "list" {
		return with_go.List()
	}
	args = args[1:]
	if strings.HasPrefix(arg0, "GOROOT=") {
		goroot = strings.TrimSpace(strings.TrimPrefix(arg0, "GOROOT="))
		if goroot == "" {
			return errors.New("example: kool with-go GOROOT=<X> ...")
		}
	} else {
		goVersion := "go" + strings.TrimPrefix(arg0, "go")
		if goVersion == "" {
			return errors.New("example: kool with-go go1.18 ...")
		}
		goroot, err = resolveGoroot(goVersion)
		if err != nil {
			return err
		}
	}
	return execGoroot(goroot, args)
}

func handleWithGoroot(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("example: kool with-goroot <GOROOT>")
	}
	return execGoroot(args[0], args[1:])
}

func resolveGoroot(goVersion string) (string, error) {
	switch goVersion {
	case "go1.24":
		goVersion = "go1.24.1"
	case "go1.23":
		goVersion = "go1.23.6"
	case "go1.22":
		goVersion = "go1.22.12"
	case "go1.21":
		goVersion = "go1.21.13"
	case "go1.20":
		goVersion = "go1.20.14"
	case "go1.19":
		goVersion = "go1.19.13"
	case "go1.18":
		goVersion = "go1.18.10"
	case "go1.17":
		goVersion = "go1.17.13"
	}
	return with_go.InstallGo(goVersion, "")
}

func execGoroot(goroot string, args []string) error {
	absGoroot, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}
	envs := os.Environ()
	PATH := filepath.Join(absGoroot, "bin") + string(os.PathListSeparator) + os.Getenv("PATH")
	envs = append(envs, "GOROOT="+absGoroot, "PATH="+PATH)

	if len(args) >= 2 && args[0] == "go" && args[1] == "run" {
		err := os.Setenv("PATH", PATH)
		if err != nil {
			return err
		}
		err = os.Setenv("GOROOT", absGoroot)
		if err != nil {
			return err
		}

		// use kool go
		return run.Handle(args[2:])
	}

	var targetCmd string
	var targetArgs []string
	if len(args) == 0 {
		targetCmd = "env"
	} else {
		targetCmd = args[0]
		targetArgs = args[1:]

		strip := strings.TrimPrefix(targetCmd, "./")
		if strip == filepath.Base(targetCmd) {
			// try lookup in $GOROOT/bin
			fullCmd := filepath.Join(absGoroot, "bin", targetCmd)
			if _, err := os.Stat(fullCmd); err == nil {
				targetCmd = fullCmd
			}
		}
	}

	execCmd := exec.Command(targetCmd, targetArgs...)
	execCmd.Env = envs
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}

type Topic struct {
	Title       string
	Keywords    []string
	Description string
	SubTopics   []*Topic
}

var topics = []*Topic{
	{
		Title:       "mermaid",
		Keywords:    []string{"mermaid"},
		Description: "mermaid to image",
		SubTopics: []*Topic{
			{
				Title:    "mermaid to image",
				Keywords: []string{"mermaid to image"},
				Description: `# install mmdc
npm install -g @mermaid-js/mermaid-cli

# set a large width so resolution
# won't compromise
mmdc -i input.mmd -o output.png --width 10000

# on MacOS, paste from clipboard
mmdc -i <(pbpaste) -o output.png --width 10000

check https://github.com/mermaid-js/mermaid-cli
`,
			},
		},
	},
}

func traverseTopics(topic *Topic, unit string, indent string) {
	fmt.Printf("%s- %s\n", indent, topic.Title)
	nextIndent := unit + indent
	for _, subTopic := range topic.SubTopics {
		traverseTopics(subTopic, unit, nextIndent)
	}
}

func handleQuestion(args []string) error {
	if len(args) == 0 {
		for _, topic := range topics {
			traverseTopics(topic, "  ", "")
		}
		return nil
	}

	question := strings.Join(args, " ")

	var answerQuestion func(topic *Topic)
	answerQuestion = func(topic *Topic) {
		if topic.Title == question {
			fmt.Printf(topic.Description)
		}
		for _, subTopic := range topic.SubTopics {
			answerQuestion(subTopic)
		}
	}

	for _, topic := range topics {
		answerQuestion(topic)
	}

	return nil
}
