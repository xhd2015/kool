package main

import (
	"fmt"

	"github.com/xhd2015/kool/tools/git_tag_next"
)

func handleGit(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("commands: tag-next")
	}
	switch args[0] {
	case "tag-next":
		return git_tag_next.Handle(args[1:])
	}
	return fmt.Errorf("unknown command: %s", args[0])
}
