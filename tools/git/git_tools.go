package git

import (
	"fmt"
	"strings"

	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/kool/tools/git/check_merged"
	"github.com/xhd2015/kool/tools/git/git_check_merge"
	"github.com/xhd2015/kool/tools/git/git_show_exclude"
	"github.com/xhd2015/kool/tools/git/git_tag_next"
	"github.com/xhd2015/kool/tools/git/grep"
	"github.com/xhd2015/kool/tools/git/hooks"
	"github.com/xhd2015/kool/tools/git/ls"
	"github.com/xhd2015/kool/tools/git/staged"
	"github.com/xhd2015/kool/tools/git/worktree"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
kool git enhances the git command line tools.

Usage: kool git <cmd> [OPTIONS]

Available commands:
  ls                               list files that is able to be committed with git add -A
  worktree                         worktree commands
  tag-next                         tag next version
  show-tag                         show tag of current commit
  show-exclude                     show exclude files
  init-hooks                       init git hooks
  help                             show help message

Options:
  --dir <dir>                      set the output directory
  -v,--verbose                     show verbose info

Options for tag-next:
  --push                           tag and push to remote

Examples:
  kool git help                    show help message
  kool git ls                      list files
  kool git tag-next --push
`

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("commands: tag-next, show-tag, show-exclude, check-merge,help")
	}
	var isShowTag bool
	var remainArgs []string
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Println(strings.TrimPrefix(help, "\n"))
		return nil
	case "worktree":
		return worktree.Handle(args[1:])
	case "tag-next":
		return git_tag_next.Handle(args[1:])
	case "show-tag":
		isShowTag = true
		remainArgs = args[1:]
	case "show-exclude":
		return git_show_exclude.Handle()
	case "init-hooks":
		return hooks.HandleInit(args[1:])
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
			return HandleShowChildren(args[2:])
		default:
			return fmt.Errorf("unknown show subcommand: %s", args[1])
		}
	case "show-children":
		return HandleShowChildren(args[1:])
	case "check-merged":
		return check_merged.Handle(args[1:])
	case "check-merge":
		return git_check_merge.Handle(args[1:])
	case "ls":
		return HandleLs(args[1:])
	case "staged":
		return staged.Handle(args[1:])
	case "grep":
		return grep.Handle(args[1:])
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

func HandleShowChildren(args []string) error {
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

const lsHelp = `
Usage: kool git ls [OPTIONS]

Options:
  -h,--help            show help message
  -v,--verbose         show verbose info
  --cached
  --staged             list staged/cached files, same thing
`

func HandleLs(args []string) error {
	var dir string
	var staged bool
	var verbose bool
	args, err := flags.String("--dir", &dir).
		Bool("--cached,--staged", &staged).
		Bool("-v,--verbose", &verbose).
		Help("-h,--help", lsHelp).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	if staged {
		return ls.LsStagedFiles(dir, verbose)
	}

	return lsCommitFiles(dir)
}

func lsCommitFiles(dir string) error {
	seen := make(map[string]bool)

	printFile := func(file string) {
		if seen[file] {
			return
		}
		seen[file] = true
		fmt.Println(file)
	}

	addedFiles, err := git.ListAddedFile(dir, git.COMMIT_WORKING, "HEAD", nil)
	if err != nil {
		return err
	}
	for _, file := range addedFiles {
		printFile(file)
	}

	files, err := git.ListModifiedFiles(dir, git.COMMIT_WORKING, "HEAD", nil)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		printFile(file)
	}

	renamedFiles, err := git.ListRenamedFiles(dir, git.COMMIT_WORKING, "HEAD", nil)
	if err != nil {
		return err
	}
	for _, file := range renamedFiles {
		printFile(file.File)
	}

	untrackedFiles, err := git.ListUntrackedFiles(dir, "HEAD", nil)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range untrackedFiles {
		printFile(file)
	}
	return nil
}
