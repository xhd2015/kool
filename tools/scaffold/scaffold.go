package scaffold

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: kool scaffold [--list]
       kool scaffold <name>

Scaffolds:
  go-cmd-run-lib

Options:
  --list       list available scaffold names
  -h,--help    show help message
`

type scaffold struct {
	Name    string
	Content string
}

var scaffolds = []scaffold{
	{
		Name:    "go-cmd-run-lib",
		Content: goCmdRunLibScaffold,
	},
}

const goCmdRunLibScaffold = `# cmd/__NAME__/main.go
package main

import (
	"fmt"
	"os"

	__NAME__ "__MODULE__/run/__NAME__"
)

func main() {
	if err := __NAME__.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "__NAME__: %v\n", err)
		os.Exit(1)
	}
}

# run/__NAME__/run.go
package __NAME__

import (
	"fmt"
	"strings"

	core "__MODULE__/pkgs/__NAME__"
	"github.com/xhd2015/less-gen/flags"
)

const help = ` + "`" + `
Usage: __NAME__ [OPTIONS]

Options:
  -h,--help    show help message
` + "`" + `

func Run(args []string) error {
	config := core.Config{}
	args, err := flags.
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	return core.Run(config)
}

# pkgs/__NAME__/__NAME__.go
package __NAME__

type Config struct{}

func Run(config Config) error {
	// Core library logic goes here.
	return nil
}
`

func Handle(args []string) error {
	return HandleWithWriter(os.Stdout, args)
}

func HandleWithWriter(w io.Writer, args []string) error {
	var list bool
	args, err := flags.Bool("--list", &list).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if list {
		if len(args) > 0 {
			return fmt.Errorf("--list does not accept arguments")
		}
		for _, scaffold := range scaffolds {
			fmt.Fprintln(w, scaffold.Name)
		}
		return nil
	}

	if len(args) == 0 {
		fmt.Fprint(w, strings.TrimPrefix(help, "\n"))
		return nil
	}
	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra arguments: %s", strings.Join(args[1:], " "))
	}

	content, ok := lookup(args[0])
	if !ok {
		return fmt.Errorf("unknown scaffold: %s", args[0])
	}
	fmt.Fprint(w, content)
	return nil
}

func lookup(name string) (string, bool) {
	for _, scaffold := range scaffolds {
		if scaffold.Name == name {
			return scaffold.Content, true
		}
	}
	return "", false
}
