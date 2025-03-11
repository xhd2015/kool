package git_tag_next

import (
	"os"
	"os/exec"
	"strings"
)

// use `git describe --tags HEAD` to show the tag
// NOTE: though `git tag -l --points-at HEAD` works for tagged commits,
// we use `git describe --tags HEAD`, which works for both tagged and untagged commits
func ShowHeadTag(dir string) (string, error) {
	gitDescribe := exec.Command("git", "describe", "--tags", "HEAD")
	gitDescribe.Dir = dir
	gitDescribe.Stderr = os.Stderr
	output, err := gitDescribe.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
