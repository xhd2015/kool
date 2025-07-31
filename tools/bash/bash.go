package bash

import (
	"fmt"

	"github.com/xhd2015/kool/tools/bash/history"
)

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires command: history")
	}

	cmd := args[0]
	args = args[1:]

	if cmd == "history" {
		return history.Handle(args)
	}

	return fmt.Errorf("unknown command: %s", cmd)
}
