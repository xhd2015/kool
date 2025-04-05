package main

import (
	"fmt"

	"github.com/xhd2015/kool/tools/git_show_exclude"
	"github.com/xhd2015/kool/tools/git_tag_next"
)

func handleGit(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("commands: tag-next, show-tag, show exclude")
	}
	var isShowTag bool
	var remainArgs []string
	switch args[0] {
	case "tag-next":
		return git_tag_next.Handle(args[1:])
	case "show-tag":
		isShowTag = true
		remainArgs = args[1:]
	case "show":
		if len(args) < 2 {
			return fmt.Errorf("expected subcommand for show: exclude,tag,children")
		}
		if args[1] == "tag" {
			isShowTag = true
			remainArgs = args[2:]
			break
		}
		switch args[1] {
		case "exclude":
			return git_show_exclude.Handle()
		case "children":
			return handleGitShowChildren(args[2:])
		default:
			return fmt.Errorf("unknown show subcommand: %s", args[1])
		}
	case "show-children":
		return handleGitShowChildren(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}

	if isShowTag {
		var dir string
		if len(remainArgs) > 0 {
			dir = remainArgs[0]
		}
		tag, err := git_tag_next.ShowHeadTag(dir)
		if err != nil {
			return err
		}
		fmt.Println(tag)
	}
	return nil
}

func handleGitShowChildren(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: git show-children [commit hash]")
	}
	if len(args) > 1 {
		return fmt.Errorf("expected only one commit hash")
	}
	commit := args[0]
	children, err := git_tag_next.ShowChildren("", commit)
	if err != nil {
		return err
	}
	for _, child := range children {
		fmt.Println(child)
	}
	return nil
}
