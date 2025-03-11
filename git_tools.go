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
	case "show-tag":
		var dir string
		if len(args) > 1 {
			dir = args[1]
		}
		tag, err := git_tag_next.ShowHeadTag(dir)
		if err != nil {
			return err
		}
		fmt.Println(tag)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
	return nil
}
