package ls

import (
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

func LsStagedFiles(dir string, verbose bool) error {
	c := cmd.New()
	if verbose {
		c.Debug()
	}

	return c.Dir(dir).Run("git", "diff", "--name-only", "--cached")
}

func GetStatedFiles(dir string, verbose bool) ([]string, error) {
	c := cmd.New()
	if verbose {
		c.Debug()
	}
	// exclude deleted files
	output, err := c.Dir(dir).Output("git", "diff", "--name-only", "--cached", "--diff-filter=ACMR")
	if err != nil {
		return nil, err
	}
	return toList(output), nil
}

func toList(s string) []string {
	list := strings.Split(s, "\n")
	result := make([]string, 0, len(list))
	for i := range list {
		e := strings.TrimSpace(list[i])
		if e == "" {
			continue
		}
		result = append(result, e)
	}
	return result
}
