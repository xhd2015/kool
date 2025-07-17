package worktree

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

const help = `
kool git worktree facility

Usage: kool git worktree <cmd> [OPTIONS]

Available commands:
  add <path> [branch]              add a new worktree, with optional setting new branch
  help                             show help message

Options:
  --dir <dir>                      set the output directory
  -v,--verbose                     show verbose info  

Examples:
  kool git worktree help                           show help message
  kool git worktree add ../working-v1.2.0 v1.2.0   create a new project named my_project
`

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires subcommands: help, add")
	}
	commd := args[0]
	args = args[1:]
	switch commd {
	case "add":
		return add(args)
	case "help", "--help", "-h":
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
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
		err := cmd.Debug().Dir(pathName).Run("git", "reset", "--hard", "origin/"+branch)
		if err != nil {
			return err
		}
		err = cmd.Debug().Dir(pathName).Run("git", "branch", "-u", "origin/"+branch)
		if err != nil {
			return err
		}
	}

	return nil
}
