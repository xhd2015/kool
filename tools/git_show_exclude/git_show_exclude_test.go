package git_show_exclude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetExcludePath(t *testing.T) {
	// This test assumes it's being run in a git repository
	path, err := GetExcludePath()
	if err != nil {
		t.Fatalf("Error getting exclude path: %v", err)
	}

	// Check that the path refers to a .git/info/exclude file
	if !strings.Contains(path, "info/exclude") && !strings.Contains(path, "info\\exclude") {
		t.Errorf("Path doesn't point to info/exclude: %s", path)
	}
}

func TestRelativePathLogic(t *testing.T) {
	// Create a mock directory structure for testing
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock exclude file
	gitDir := filepath.Join(tmpDir, ".git")
	infoDir := filepath.Join(gitDir, "info")
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		t.Fatalf("Failed to create info dir: %v", err)
	}
	excludePath := filepath.Join(infoDir, "exclude")
	if err := os.WriteFile(excludePath, []byte("# Test exclude file"), 0644); err != nil {
		t.Fatalf("Failed to create exclude file: %v", err)
	}

	// Test relative path counting
	parts := strings.Split("../../.git/info/exclude", "/")
	dotDotCount := 0
	for _, part := range parts {
		if part == ".." {
			dotDotCount++
		}
	}
	if dotDotCount != 2 {
		t.Errorf("Expected 2 .., got %d", dotDotCount)
	}
}
