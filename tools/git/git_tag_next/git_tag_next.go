package git_tag_next

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const help = `
usage: git-tag-next [OPTIONS]

Options:
   --show       show the tag, no creating or pushing
   --push       push the tag
   --help,-h    help

Example:
    $ git-tag-next --show
`

type Options struct {
	Show bool
	Push bool
}

func Handle(args []string) error {
	opts := &Options{}
	// Parse options
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--show":
			opts.Show = true
		case "--push":
			opts.Push = true
		case "--help", "-h":
			fmt.Println(strings.TrimPrefix(help, "\n"))
			return nil
		default:
			return fmt.Errorf("unknown option: %s", args[i])
		}
	}

	return handleGitTag(opts)
}

func handleGitTag(opts *Options) error {
	// Check if HEAD is already tagged
	start := 0
	if !opts.Show {
		start = 1
		headTag, err := execGit("tag", "-l", "--points-at", "HEAD")
		if err != nil {
			return err
		}
		if headTag != "" {
			return fmt.Errorf("already tagged: %s", headTag)
		}
	}

	// Find most recent tag
	var tag string
	for i := start; i < start+10; i++ {
		commit := fmt.Sprintf("HEAD~%d", i)
		currentTag, err := execGit("tag", "-l", "--points-at", commit)
		if err != nil {
			return err
		}
		if currentTag != "" {
			tag = currentTag
			break
		}
	}

	if tag == "" {
		return fmt.Errorf("no tag for recent 10 commits, please make a new one manually")
	}

	if strings.Contains(tag, "/") {
		return fmt.Errorf("tag contains '/', please make a new one manually: %s", tag)
	}

	// Calculate next tag
	nextTag, err := incrementTag(tag)
	if err != nil {
		return fmt.Errorf("failed to increment tag: %v", err)
	}

	if opts.Show {
		fmt.Println(nextTag)
		return nil
	}

	// Create new tag
	if err := runGitCmdPipeOutput("tag", nextTag); err != nil {
		return fmt.Errorf("failed to create tag: %v", err)
	}

	if opts.Push {
		// Get current branch
		branch, err := execGit("branch", "--show-current")
		if err != nil {
			return fmt.Errorf("failed to get current branch: %v", err)
		}

		// Push changes
		if err := runGitCmdPipeOutput("push", "origin", fmt.Sprintf("HEAD:%s", branch)); err != nil {
			return fmt.Errorf("failed to push changes: %v", err)
		}

		if err := runGitCmdPipeOutput("push", "origin", nextTag); err != nil {
			return fmt.Errorf("failed to push tag: %v", err)
		}
	}

	return nil
}

func incrementTag(tag string) (string, error) {
	// Find the last numeric part
	n := len(tag)
	i := n - 1
	for ; i >= 0; i-- {
		if tag[i] < '0' || tag[i] > '9' {
			break
		}
	}

	// Check leading zeros
	for ; i < n-1; i++ {
		if tag[i+1] != '0' {
			break
		}
	}

	if i >= n-1 {
		return "", fmt.Errorf("non numeric tag: %s", tag)
	}

	numStr := tag[i+1:]
	num, err := strconv.Atoi(numStr)
	if err != nil || num <= 0 {
		return "", fmt.Errorf("invalid tag: %s", tag)
	}

	return tag[:i+1] + strconv.Itoa(num+1), nil
}

func execGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(output)), nil
}

func runGitCmdPipeOutput(args ...string) error {
	fmt.Println("git", strings.Join(args, " "))
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s failed: %v", strings.Join(args, " "), err)
	}
	return nil
}
