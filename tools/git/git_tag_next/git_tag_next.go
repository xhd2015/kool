package git_tag_next

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/kool/tools/git/tag"
	"github.com/xhd2015/less-gen/flags"
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
	Dir     string
	Show    bool
	Push    bool
	Verbose bool
}

func Handle(args []string) error {
	opts := Options{}
	args, err := flags.String("--dir", &opts.Dir).
		Bool("--show", &opts.Show).
		Bool("--push", &opts.Push).
		Help("-h,--help", help).
		Bool("-v,--verbose", &opts.Verbose).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	return handleGitTag(opts)
}

func handleGitTag(opts Options) error {
	dir := opts.Dir
	versionPrefix, err := tag.GetVersionPrefix(dir)
	if err != nil {
		return err
	}

	// Find most recent latestTag
	latestTag, err := tag.GetLatestVersionTag(dir, versionPrefix)
	if err != nil {
		return fmt.Errorf("failed to get latest version tag: %w", err)
	}

	if latestTag == "" {
		return fmt.Errorf("no latest version tag found, please make a new one manually")
	}

	latestTagCommit, err := git.RevParseVerified(dir, latestTag)
	if err != nil {
		return fmt.Errorf("failed to get latest version tag commit: %w", err)
	}
	headCommit, err := git.RevParseVerified(dir, "HEAD")
	if err != nil {
		return fmt.Errorf("failed to get head commit: %w", err)
	}

	if latestTagCommit == headCommit {
		if opts.Show {
			fmt.Println(latestTag)
			return nil
		}
		return fmt.Errorf("latest version tag %s is already tagged at HEAD", latestTag)
	}

	// Calculate next tag
	nextTag, err := incrementTag(latestTag)
	if err != nil {
		return fmt.Errorf("failed to increment tag: %v", err)
	}

	if opts.Show {
		fmt.Println(nextTag)
		return nil
	}

	// Create new tag
	if err := runGitCmdPipeOutput(dir, "tag", nextTag); err != nil {
		return fmt.Errorf("failed to create tag: %v", err)
	}

	if opts.Push {
		// Get current branch
		branch, err := execGit(dir, "branch", "--show-current")
		if err != nil {
			return fmt.Errorf("failed to get current branch: %v", err)
		}

		// Push changes
		if err := runGitCmdPipeOutput(dir, "push", "origin", fmt.Sprintf("HEAD:%s", branch)); err != nil {
			return fmt.Errorf("failed to push changes: %v", err)
		}

		if err := runGitCmdPipeOutput(dir, "push", "origin", nextTag); err != nil {
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

func execGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(output)), nil
}

func runGitCmdPipeOutput(dir string, args ...string) error {
	fmt.Println("git", strings.Join(args, " "))
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s failed: %v", strings.Join(args, " "), err)
	}
	return nil
}
