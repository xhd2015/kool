package git_tag_next

import (
	"os"
	"os/exec"
	"strings"
)

func ShowCurrentBranch(dir string) (string, error) {
	gitBranch := exec.Command("git", "branch", "--show-current")
	gitBranch.Dir = dir
	gitBranch.Stderr = os.Stderr

	output, err := gitBranch.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
