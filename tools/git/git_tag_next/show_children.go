package git_tag_next

// credit to https://gist.github.com/kohsuke/7590246

// #!/bin/bash -e
// # given a commit, find immediate children of that commit.
// for arg in "$@"; do
//
//	for commit in $(git rev-parse $arg^0); do
//	  for child in $(git log --format='%H %P' --all | grep -F " $commit" | cut -f1 -d' '); do
//	    git describe $child
//	  done
//	done
//
// done

import (
	"fmt"
	"os/exec"
	"strings"
)

// reference: https://git-scm.com/docs/pretty-formats
// %H: commit hash
// %P: parent commit hash
func ShowChildren(dir string, commit string) ([]string, error) {
	if commit == "" {
		return nil, fmt.Errorf("required commit hash")
	}
	// First normalize the commit hash
	normalizedCmd := exec.Command("git", "rev-parse", commit+"^0")
	normalizedCmd.Dir = dir
	normalizedBytes, err := normalizedCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to normalize commit: %w", err)
	}

	normalizedCommit := strings.TrimSpace(string(normalizedBytes))

	// Get all commits with their parents
	logCmd := exec.Command("git", "log", "--format=%H %P", "--all")
	logCmd.Dir = dir
	logBytes, err := logCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	// Parse the output to find child commits
	var children []string
	logLines := strings.Split(string(logBytes), "\n")
	for _, line := range logLines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			continue
		}
		p := strings.TrimSpace(parts[1])
		if p != normalizedCommit {
			continue
		}
		h := strings.TrimSpace(parts[0])
		if h == "" {
			continue
		}

		children = append(children, h)
	}

	return children, nil
}
