package for_each_dir

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/less-flags"
)

const help = `
kool for-each-dir - Run a command in each subdirectory

Usage: kool for-each-dir [dir] <command> [args...]

Arguments:
  dir                              directory to list subdirectories from (default: .)
  command                          command to run in each subdirectory
  args                             arguments for the command

Options:
  -h,--help                        show help message

Examples:
  kool for-each-dir ./cmd my-cli do-something with args
  kool for-each-dir . ls -la
`

func Handle(args []string) error {
	args, err := lessflags.Help("-h,--help", help).StopOnFirstArg().Parse(args)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return fmt.Errorf("requires command: kool for-each-dir [dir] <command> [args...]")
	}
	var dir string
	var command string
	var cmdArgs []string
	stat, err := os.Stat(args[0])
	if err == nil && stat.IsDir() {
		dir = args[0]
		if len(args) < 2 {
			return fmt.Errorf("requires command: kool for-each-dir [dir] <command> [args...]")
		}
		command = args[1]
		cmdArgs = args[2:]
	} else {
		dir = "."
		command = args[0]
		cmdArgs = args[1:]
	}
	return forEachDir(dir, command, cmdArgs)
}

func forEachDir(dir string, command string, cmdArgs []string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", dir, err)
	}
	var lastErr error
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(dir, entry.Name())
		cmd := exec.Command(command, cmdArgs...)
		cmd.Dir = subDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			lastErr = fmt.Errorf("command failed in %s: %v", subDir, err)
			fmt.Fprintf(os.Stderr, "error: %v\n", lastErr)
		}
	}
	return lastErr
}
