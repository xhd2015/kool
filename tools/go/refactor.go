package go_tools

import (
	"fmt"

	"github.com/xhd2015/kool/tools/go/move"
	"github.com/xhd2015/less-gen/flags"
)

const refactorHelp = `
kool go refactor helps to refactor go code.

Usage: kool go refactor <command> [OPTIONS]

Commands:
  move <src> <dst>    move a package and rewrite imports

Run kool go refactor <command> --help for more information.
`

func HandleRefactor(args []string) error {
	args, err := flags.
		Help("-h,--help", refactorHelp).
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: kool go refactor <command>")
	}
	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "move":
		return move.Handle(args)
	}
	return fmt.Errorf("unknown command: %s", cmd)
}
