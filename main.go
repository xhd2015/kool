package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/kool/tools/ai"
	"github.com/xhd2015/kool/tools/bash"
	"github.com/xhd2015/kool/tools/create"
	"github.com/xhd2015/kool/tools/debug"
	"github.com/xhd2015/kool/tools/dlv"
	"github.com/xhd2015/kool/tools/encoding"
	"github.com/xhd2015/kool/tools/git"
	go_tools "github.com/xhd2015/kool/tools/go"
	"github.com/xhd2015/kool/tools/go/with_go"
	"github.com/xhd2015/kool/tools/html/html2markdown"
	"github.com/xhd2015/kool/tools/html/html2text"
	"github.com/xhd2015/kool/tools/http"
	"github.com/xhd2015/kool/tools/js"
	"github.com/xhd2015/kool/tools/json2yaml"
	"github.com/xhd2015/kool/tools/jsontool"
	"github.com/xhd2015/kool/tools/port"
	"github.com/xhd2015/kool/tools/preview"
	"github.com/xhd2015/kool/tools/react"
	"github.com/xhd2015/kool/tools/rules"
	"github.com/xhd2015/kool/tools/stringtool"
	"github.com/xhd2015/kool/tools/uuid"
	"github.com/xhd2015/kool/tools/watch"
	"github.com/xhd2015/kool/tools/yaml2json"
	xgo_cmd "github.com/xhd2015/xgo/support/cmd"
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
  watch <command> [args...]      watch files and restart command on changes
  preview <file>                     preview a file, currently supports .uml and .puml
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
	if len(args) == 0 {
		// to suppress lint warning
		var DOT = "."
		return errors.New("requires command, try 'kool --help'" + DOT)
	}
	cmd := args[0]
	if cmd == "help" || cmd == "--help" || cmd == "-h" {
		fmt.Println(strings.TrimSpace(strings.ReplaceAll(help, "\t", "    ")))
		return nil
	}
	args = args[1:]
	switch cmd {
	case "upgrade":
		return xgo_cmd.Debug().Run("go", "install", "github.com/xhd2015/kool@latest")
	case "vscode":
		return handleVscode(args)
	case "create":
		return create.Handle(args)
	case "snippet":
		return handleSnippet(args)
	case "debug":
		return debug.Handle(args)
	case "go":
		return go_tools.Handle(args)
	case "go-replace":
		return go_tools.HandleReplace(args)
	case "go-update":
		return go_tools.HandleUpdate(args)
	case "go-resolve":
		return go_tools.HandleResolve(args)
	case "dlv":
		return dlv.Handle(args)
	case "git":
		return git.Handle(args)
	case "http":
		return http.Handle(args)
	case "with":
		return handleWith(args)
	case "with-go":
		return with_go.Handle(args, nil)
	case "with-goroot":
		return with_go.HandleWithGoroot(args, nil)
	case "ai":
		return ai.Handle(args)
	case "rule", "rules":
		return rules.Handle(args)
	case "check-port-ready":
		return port.CheckReady(args)
	case "react":
		return react.Handle(args)
	case "preview":
		return preview.Handle(args)
	case "kill-port":
		return port.HandleKill(args)
	case "watch":
		return watch.Handle(args)
		// strings
	case "lines":
		return stringtool.HandleLines(args)
	case "strcase":
		return stringtool.HandleStrCase(args)
	case "quote":
		return stringtool.HandleQuote(args)
	case "unquote":
		return stringtool.HandleUnquote(args)
	case "split":
		return stringtool.HandleSplit(args)
	case "decode":
		return encoding.HandleDecode(args)
	case "encode":
		return encoding.HandleEncode(args)
	case "bash":
		return bash.Handle(args)
		// jsons
	case "sample":
		return jsontool.HandleSample(args)
	case "pretty":
		return jsontool.HandlePretty(args)
	case "compress":
		return jsontool.HandleCompress(args)
	case "yaml2json", "yml2json":
		return yaml2json.Handle(args)
	case "json2yaml", "json2yml":
		return json2yaml.Handle(args)
	case "html2text":
		return html2text.Handle(args)
	case "html2md", "html2markdown":
		return html2markdown.Handle(args)
	case "uuid":
		return uuid.Handle(args)
	case "js":
		return js.Handle(args)
	case "?":
		return handleQuestion(args)
	default:
		if strings.HasPrefix(cmd, "with-") {
			withCmd := strings.TrimPrefix(cmd, "with-")
			if withCmd == "" {
				return fmt.Errorf("example: kool with-go1.23")
			}
			return handleWithCmd(withCmd, args)
		}

		// capture unknown command
		if cmd != "" {
			return fmt.Errorf("unrecognized command: %s", cmd)
		}
	}
	return nil
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
		goroot, err := with_go.ResolveGoroot(cmd)
		if err != nil {
			return err
		}
		return with_go.ExecGoroot(goroot, args, nil)
	}
	return fmt.Errorf("unknown command: %s", cmd)
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
	{
		Title:    "cursor",
		Keywords: []string{"cursor", "ide"},
		Description: `# cursor history
~/Library/Application Support/Cursor/User/History
`,
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
