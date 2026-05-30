package git_tmp_exclude

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const marker = "# kool-git-tmp-ignore"
const help = `
Usage: kool git tmp-exclude <pattern> [<pattern>...]
       kool git tmp-ignore  <pattern> [<pattern>...]

  Temporarily add patterns to .git/info/exclude.
  Each pattern is prefixed with a marker comment for
  easy identification later.

Options:
  -h,--help   show help message
`

func Handle(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Println(strings.TrimPrefix(help, "\n"))
		return nil
	}

	excludePath, err := getExcludePath()
	if err != nil {
		return fmt.Errorf("failed to get exclude path: %w", err)
	}

	var content []byte
	if data, err := os.ReadFile(excludePath); err == nil {
		content = data
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read exclude file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	var added []string
	for _, pattern := range args {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if hasEntry(lines, pattern) {
			continue
		}
		lines = append(lines, marker, pattern, "")
		added = append(added, pattern)
	}

	if len(added) == 0 {
		fmt.Printf("no new patterns to add to %s\n", excludePath)
		return nil
	}

	newContent := strings.Join(lines, "\n")
	if err := os.MkdirAll(filepath.Dir(excludePath), 0755); err != nil {
		return fmt.Errorf("failed to ensure info dir: %w", err)
	}
	if err := os.WriteFile(excludePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write exclude file: %w", err)
	}

	for _, p := range added {
		fmt.Printf("added %q to %s\n", p, excludePath)
	}
	return nil
}

func getExcludePath() (string, error) {
	gitDir, err := getGitCommonDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(gitDir, "info", "exclude"), nil
}

func getGitCommonDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	gitDir := strings.TrimSpace(string(output))
	if !filepath.IsAbs(gitDir) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		gitDir = filepath.Join(cwd, gitDir)
	}
	return gitDir, nil
}

func hasEntry(lines []string, pattern string) bool {
	for _, line := range lines {
		if strings.TrimSpace(line) == pattern {
			return true
		}
	}
	return false
}
