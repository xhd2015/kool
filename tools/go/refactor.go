package go_tools

import (
	"fmt"

	"github.com/xhd2015/kool/tools/go/move"
)

func HandleRefactor(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kool go refactor <command>")
	}
	switch args[0] {
	case "move":
		return move.Handle(args[1:])
	}
	return fmt.Errorf("unknown command: %s", args[0])
}
