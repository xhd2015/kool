//go:build ignore
// +build ignore

package go_tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
cli help to parse flags

Usage: cli <cmd> [OPTIONS]

Available commands:
  create <name>                    create a new project
  help                             show help message

Options:
  --dir <dir>                      set the output directory
  -v,--verbose                     show verbose info  

Examples:
  cli help                         show help message
  cli create my_project            create a new project named my_project
`

func main() {
	err := handle(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	n := len(args)
	var remainArgs []string
	var verbose bool
	var dir string
	for i := 0; i < n; i++ {
		flag, value := flags.ParseIndex(args, &i)
		if flag == "" {
			remainArgs = append(remainArgs, args[i])
			continue
		}
		switch flag {
		case "--dir":
			value, ok := value()
			if !ok {
				return fmt.Errorf("%s requires a value", flag)
			}
			dir = value
		case "-v", "--verbose":
			verbose = true
		case "-h", "--help":
			fmt.Print(strings.TrimPrefix(help, "\n"))
			return nil
		// ...
		default:
			return fmt.Errorf("unrecognized flag: %s", flag)
		}
	}
	_ = dir
	_ = verbose
	return nil
}
