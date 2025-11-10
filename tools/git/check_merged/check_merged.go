package check_merged

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: kool git check-merged <commit1> <commit2>

check if two commits are merged together.

It runs git merge-base --is-ancestor between two commits and output the relation.

Options:
  -h,--help            show help message
  -v,--verbose         show verbose info
`

func Handle(args []string) error {
	// "github.com/xhd2015/less-gen/flags"
	var verbose bool
	var dir string
	args, err := flags.String("--dir", &dir).
		Help("-h,--help", help).
		Bool("-v,--verbose", &verbose).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: kool git check-merged <commit1> <commit2>")
	}
	commit1 := args[0]
	commit2 := args[1]
	args = args[2:]
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	var isAncestor1 bool
	var isAncestor2 bool
	isAncestor1, err = isAncestor(dir, commit1, commit2)
	if err != nil {
		return err
	}
	isAncestor2, err = isAncestor(dir, commit2, commit1)
	if err != nil {
		return err
	}

	if !isAncestor1 && !isAncestor2 {
		return fmt.Errorf("%s and %s are diverged", commit1, commit2)
	}
	if isAncestor1 && isAncestor2 {
		// rev-parse:
		rev1, err := revParse(dir, commit1)
		if err != nil {
			return err
		}
		rev2, err := revParse(dir, commit2)
		if err != nil {
			return err
		}
		if rev1 == rev2 {
			fmt.Printf("%s and %s are identical commits to %s\n", commit1, commit2, rev1)
			return nil
		}
		fmt.Printf("%s and %s are merged, possibly identical commits\n", commit1, commit2)
		return nil
	}
	if isAncestor1 {
		fmt.Printf("%s is merged into %s\n", commit1, commit2)
		return nil
	}
	if isAncestor2 {
		fmt.Printf("%s is merged into %s\n", commit2, commit1)
		return nil
	}
	return fmt.Errorf("unrecognized error")
}

func isAncestor(dir string, commit1 string, commit2 string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", commit1, commit2)
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check if %s is ancestor of %s: %w", commit1, commit2, err)
	}
	return true, nil
}

func revParse(dir string, commit string) (string, error) {
	cmd := exec.Command("git", "rev-parse", commit)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to rev-parse %s: %w", commit, err)
	}
	return strings.TrimSpace(string(output)), nil
}
