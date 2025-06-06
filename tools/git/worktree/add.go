package worktree

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires subcommands: add")
	}
	commd := args[0]
	args = args[1:]
	switch commd {
	case "add":
		return add(args)
	}
	return fmt.Errorf("unrecognized subcommand: %s", commd)
}

func add(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires <path> [branch]")
	}

	pathName := args[0]
	args = args[1:]
	var branch string
	if len(args) > 0 {
		branch = args[0]
		args = args[1:]
	}

	if len(args) > 0 {
		return fmt.Errorf("unexpected extra arguments: %v", strings.Join(args, ","))
	}

	if pathName == "" {
		return fmt.Errorf("requries pathName")
	}

	gitArgs := []string{
		"worktree", "add", pathName,
	}

	if branch != "" {
		gitArgs = append(gitArgs, "-B", branch)
	}

	err := cmd.Debug().Run("git", gitArgs...)
	if err != nil {
		return err
	}

	if branch != "" {
		err := cmd.Debug().Dir(pathName).Run("git", "branch", "-u", "origin/"+branch)
		if err != nil {
			return err
		}
	}

	return nil
}
