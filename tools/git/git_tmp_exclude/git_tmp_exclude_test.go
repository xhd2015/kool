package git_tmp_exclude

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTmpExclude(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tmp-exclude-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	if err := initTestRepo(testDir); err != nil {
		t.Fatalf("Failed to init test repo: %v", err)
	}

	t.Run("adds pattern to exclude", func(t *testing.T) {
		if err := runInDir(testDir, func() error {
			return Handle([]string{"*.log"})
		}); err != nil {
			t.Fatalf("Handle failed: %v", err)
		}

		excludePath := filepath.Join(testDir, ".git", "info", "exclude")
		content := readFile(t, excludePath)
		if !strings.Contains(content, marker) {
			t.Errorf("expected marker %q in exclude file, got:\n%s", marker, content)
		}
		if !strings.Contains(content, "*.log") {
			t.Errorf("expected pattern *.log in exclude file, got:\n%s", content)
		}
	})

	t.Run("dedup skips existing pattern", func(t *testing.T) {
		excludePath := filepath.Join(testDir, ".git", "info", "exclude")
		content := readFile(t, excludePath)
		markerCount := strings.Count(content, marker)
		starLogCount := strings.Count(content, "*.log")

		if err := runInDir(testDir, func() error {
			return Handle([]string{"*.log"})
		}); err != nil {
			t.Fatalf("Handle failed: %v", err)
		}

		content2 := readFile(t, excludePath)
		if strings.Count(content2, marker) != markerCount {
			t.Errorf("expected %d marker occurrences, got %d", markerCount, strings.Count(content2, marker))
		}
		if strings.Count(content2, "*.log") != starLogCount {
			t.Errorf("expected %d *.log occurrences, got %d", starLogCount, strings.Count(content2, "*.log"))
		}
	})

	t.Run("adds multiple patterns", func(t *testing.T) {
		if err := runInDir(testDir, func() error {
			return Handle([]string{"*.tmp", "build/"})
		}); err != nil {
			t.Fatalf("Handle failed: %v", err)
		}

		excludePath := filepath.Join(testDir, ".git", "info", "exclude")
		content := readFile(t, excludePath)
		if !strings.Contains(content, "*.tmp") {
			t.Errorf("expected pattern *.tmp in exclude file")
		}
		if !strings.Contains(content, "build/") {
			t.Errorf("expected pattern build/ in exclude file")
		}
	})
}

func TestTmpExcludeWorktree(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tmp-exclude-worktree")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mainDir := filepath.Join(tmpDir, "main")
	if err := os.MkdirAll(mainDir, 0755); err != nil {
		t.Fatalf("Failed to create main dir: %v", err)
	}
	if err := initTestRepo(mainDir); err != nil {
		t.Fatalf("Failed to init test repo: %v", err)
	}

	// create an initial commit so worktree add works
	if err := createInitialCommit(mainDir); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// create a worktree
	wtDir := filepath.Join(tmpDir, "wt")
	if err := createWorktree(mainDir, wtDir); err != nil {
		t.Fatalf("Failed to create worktree: %v", err)
	}

	t.Run("writes to main repo exclude from worktree", func(t *testing.T) {
		if err := runInDir(wtDir, func() error {
			return Handle([]string{"*.wt-log"})
		}); err != nil {
			t.Fatalf("Handle failed: %v", err)
		}

		// pattern should be in the main repo's exclude, not the worktree's
		mainExcludePath := filepath.Join(mainDir, ".git", "info", "exclude")
		content := readFile(t, mainExcludePath)
		if !strings.Contains(content, "*.wt-log") {
			t.Errorf("expected *.wt-log in main repo exclude, got:\n%s", content)
		}

		// worktree's .git is a file, not a dir, so no info/exclude there
		wtExcludePath := filepath.Join(wtDir, ".git", "info", "exclude")
		if _, err := os.Stat(wtExcludePath); err == nil {
			wtContent := readFile(t, wtExcludePath)
			if strings.Contains(wtContent, "*.wt-log") {
				t.Errorf("pattern should NOT be in worktree exclude file")
			}
		}
	})
}

func TestTmpExcludeMissingDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tmp-exclude-missing")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	if err := initTestRepo(testDir); err != nil {
		t.Fatalf("Failed to init test repo: %v", err)
	}

	// remove the info dir to test auto-creation
	infoDir := filepath.Join(testDir, ".git", "info")
	if err := os.RemoveAll(infoDir); err != nil {
		t.Fatalf("Failed to remove info dir: %v", err)
	}

	t.Run("creates info dir if missing", func(t *testing.T) {
		if err := runInDir(testDir, func() error {
			return Handle([]string{"*.build"})
		}); err != nil {
			t.Fatalf("Handle failed: %v", err)
		}

		excludePath := filepath.Join(testDir, ".git", "info", "exclude")
		content := readFile(t, excludePath)
		if !strings.Contains(content, "*.build") {
			t.Errorf("expected *.build in exclude file, got:\n%s", content)
		}
	})
}

func initTestRepo(dir string) error {
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run %v: %w", c, err)
		}
	}
	return nil
}

func createInitialCommit(dir string) error {
	// create a file and commit it so worktree add works
	file := filepath.Join(dir, "README.md")
	if err := os.WriteFile(file, []byte("# test"), 0644); err != nil {
		return err
	}
	cmds := [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "initial"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run %v: %w", c, err)
		}
	}
	return nil
}

func createWorktree(mainDir, wtDir string) error {
	cmd := exec.Command("git", "worktree", "add", wtDir)
	cmd.Dir = mainDir
	return cmd.Run()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

func runInDir(dir string, fn func() error) error {
	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(dir); err != nil {
		return err
	}
	return fn()
}
