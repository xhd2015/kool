package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/term"
)

// install: go build -o $GOPATH/bin/kool
const help = `
kool help to parse

Usage: kool <cmd> x [OPTIONS]

Available commands:
  vscode                            print example vscode configs
  vscode debug-go <prog> [args...]  print vscode config for debugging go program with args

Options:
  --help   show help message
`

// install go build -o `which kool` ./
func main() {
	err := handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	var cmd string
	if len(args) > 0 && args[0] == "sample" {
		cmd = "sample"
		args = args[1:]
	}

	if len(args) > 0 && args[0] == "vscode" {
		return handleVscode(args[1:])
	}

	var some string

	var remainArgs []string
	n := len(args)
	for i := 0; i < n; i++ {
		if args[i] == "--some" {
			if i+1 >= n {
				return fmt.Errorf("%v requires arg", args[i])
			}
			some = args[i+1]
			i++
			continue
		}
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
	// TODO handle
	_ = some

	// TTY:     go run ./
	// NOT TTY: echo yes | go run ./
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))

	if !isTTY {
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
