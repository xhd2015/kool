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
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires sub command: create")
	}
	cmd := args[0]
	args = args[1:]
	if cmd == "--help" || cmd == "help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}
	switch cmd {
	case "create":
		return handle(args)
	default:
		return fmt.Errorf("unrecognized: %s", cmd)
	}
}

func handle(args []string) error {
	var verbose bool
	var dir string
	args, err := flags.String("--dir", &dir).
		Bool("-v,--verbose", &verbose).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	_ = dir
	_ = verbose
	return nil
}
