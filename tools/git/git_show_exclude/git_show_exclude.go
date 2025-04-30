package git_show_exclude

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Handle processes the git show exclude command
func Handle() error {
	excludePath, err := GetExcludePath()
	if err != nil {
		return err
	}

	fmt.Println(excludePath)
	return nil
}

// GetExcludePath returns the path to the .git/info/exclude file
// If the path has more than one parent directory (..), it returns the absolute path
func GetExcludePath() (string, error) {
	// In Git worktrees, the exclude file is typically in the main repository's .git directory
	// Try to get the git common dir first (for worktrees this will point to the main repo's .git)
	commonDir, err := getGitCommonDir()
	if err == nil {
		// If we found a common dir, try that path first
		excludePath := filepath.Join(commonDir, "info", "exclude")
		if _, err := os.Stat(excludePath); err == nil {
			// File exists, proceed with this path
			return formatPath(excludePath)
		}
	}

	// Fall back to regular git dir approach if common dir didn't work
	gitDir, err := findGitDir()
	if err != nil {
		return "", err
	}

	// Construct the path to the exclude file
	excludePath := filepath.Join(gitDir, "info", "exclude")

	// Check if the file exists
	if _, err := os.Stat(excludePath); err != nil {
		return "", fmt.Errorf("exclude file not found: %v", err)
	}

	return formatPath(excludePath)
}

// getGitCommonDir gets the common git directory, which contains the actual exclude file
// especially important for worktrees
func getGitCommonDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	gitDir := strings.TrimSpace(string(output))

	// If path is not absolute, make it absolute
	if !filepath.IsAbs(gitDir) {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		gitDir = filepath.Join(cwd, gitDir)
	}

	return gitDir, nil
}

// formatPath formats the path to the exclude file according to requirements
// If the path has more than one parent directory (..), it returns the absolute path
func formatPath(excludePath string) (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return excludePath, nil // Return absolute path on error
	}

	// Try to get relative path
	relPath, err := filepath.Rel(cwd, excludePath)
	if err != nil {
		return excludePath, nil // Return absolute path on error
	}

	// Count the number of parent directory references (..)
	parts := strings.Split(relPath, string(filepath.Separator))
	dotDotCount := 0
	for _, part := range parts {
		if part == ".." {
			dotDotCount++
		}
	}

	// If we have more than one parent directory reference, use absolute path
	if dotDotCount > 1 {
		return excludePath, nil
	}

	return relPath, nil
}

// findGitDir finds the .git directory, handling worktree case
func findGitDir() (string, error) {
	// Try using git command first (handles worktrees properly)
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: Try to find .git by traversing up
		return findGitDirByTraversal()
	}

	gitDir := strings.TrimSpace(string(output))

	// If path is not absolute, make it absolute
	if !filepath.IsAbs(gitDir) {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		gitDir = filepath.Join(cwd, gitDir)
	}

	return gitDir, nil
}

// findGitDirByTraversal traverses up the directory tree to find the .git directory
func findGitDirByTraversal() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		gitPath := filepath.Join(cwd, ".git")
		info, err := os.Stat(gitPath)
		if err == nil {
			if info.IsDir() {
				// Regular .git directory
				return gitPath, nil
			}

			// Handle worktree case (where .git is a file)
			return readGitFileWorktree(gitPath)
		}

		// Stop at root
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}

	return "", errors.New("not a git repository (or any of the parent directories)")
}

// readGitFileWorktree reads the gitdir from a .git file in a worktree
func readGitFileWorktree(gitFilePath string) (string, error) {
	file, err := os.Open(gitFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "gitdir:") {
			gitDir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))

			// If path is not absolute, make it absolute relative to the .git file location
			if !filepath.IsAbs(gitDir) {
				gitDir = filepath.Join(filepath.Dir(gitFilePath), gitDir)
			}

			return gitDir, nil
		}
	}

	return "", errors.New("invalid .git file format")
}
