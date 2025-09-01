package staged

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStagedBackupAndRestore(t *testing.T) {
	// Create temporary directory for test git repo
	tmpDir, err := os.MkdirTemp("", "staged-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	if err := initTestRepo(tmpDir); err != nil {
		t.Fatalf("Failed to init test repo: %v", err)
	}

	// Test cases
	tests := []struct {
		name     string
		testFunc func(t *testing.T, repoDir string)
	}{
		{"TestBackupNoStagedFiles", testBackupNoStagedFiles},
		{"TestBackupRegularFiles", testBackupRegularFiles},
		{"TestBackupDeletedFiles", testBackupDeletedFiles},
		{"TestBackupRenamedFiles", testBackupRenamedFiles},
		{"TestRestoreCleanWorktree", testRestoreCleanWorktree},
		{"TestRestoreDirtyWorktree", testRestoreDirtyWorktree},
		{"TestRestoreInvalidBackup", testRestoreInvalidBackup},
		{"TestBackupRestoreRoundTrip", testBackupRestoreRoundTrip},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh test repo for each test
			testDir := filepath.Join(tmpDir, tt.name)
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}
			if err := initTestRepo(testDir); err != nil {
				t.Fatalf("Failed to init test repo: %v", err)
			}
			tt.testFunc(t, testDir)
		})
	}
}

func initTestRepo(dir string) error {
	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmd := range commands {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			return fmt.Errorf("failed to run %v: %w", cmd, err)
		}
	}
	return nil
}

func testBackupNoStagedFiles(t *testing.T, repoDir string) {
	backupFile := filepath.Join(repoDir, "backup.txt")

	// Test backup with no staged files
	err := runInDir(repoDir, func() error {
		return HandleBackup(backupFile)
	})

	if err == nil {
		t.Error("Expected error for no staged files, got nil")
	}
	if !strings.Contains(err.Error(), "no staged files found") {
		t.Errorf("Expected 'no staged files found' error, got: %v", err)
	}
}

func testBackupRegularFiles(t *testing.T, repoDir string) {
	// Create and stage some files
	file1 := filepath.Join(repoDir, "file1.txt")
	file2 := filepath.Join(repoDir, "subdir", "file2.go")

	if err := os.MkdirAll(filepath.Dir(file2), 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	if err := os.WriteFile(file1, []byte("content of file1\nline 2"), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}"), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Stage files
	if err := runGitCommand(repoDir, "add", file1, file2); err != nil {
		t.Fatalf("Failed to stage files: %v", err)
	}

	// Test backup
	backupFile := filepath.Join(repoDir, "backup.txt")
	err := runInDir(repoDir, func() error {
		return HandleBackup(backupFile)
	})

	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify backup content
	content, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	backupStr := string(content)
	if !strings.Contains(backupStr, "===== file1.txt =====") {
		t.Error("Backup missing file1.txt header")
	}
	if !strings.Contains(backupStr, "===== subdir/file2.go =====") {
		t.Error("Backup missing file2.go header")
	}
	if !strings.Contains(backupStr, "content of file1") {
		t.Error("Backup missing file1 content")
	}
	if !strings.Contains(backupStr, "package main") {
		t.Error("Backup missing file2 content")
	}
}

func testBackupDeletedFiles(t *testing.T, repoDir string) {
	// Create and commit a file
	file1 := filepath.Join(repoDir, "to-delete.txt")
	if err := os.WriteFile(file1, []byte("content to be deleted"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if err := runGitCommand(repoDir, "add", file1); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	if err := runGitCommand(repoDir, "commit", "-m", "Add file to delete"); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Delete and stage deletion
	if err := runGitCommand(repoDir, "rm", file1); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Test backup
	backupFile := filepath.Join(repoDir, "backup.txt")
	err := runInDir(repoDir, func() error {
		return HandleBackup(backupFile)
	})

	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify backup content
	content, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	backupStr := string(content)
	if !strings.Contains(backupStr, "===== to-delete.txt [Deleted] =====") {
		t.Error("Backup missing deleted file header")
	}
}

func testBackupRenamedFiles(t *testing.T, repoDir string) {
	// Create and commit a file
	oldFile := filepath.Join(repoDir, "old-name.txt")
	newFile := filepath.Join(repoDir, "new-name.txt")

	if err := os.WriteFile(oldFile, []byte("content to be renamed"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	if err := runGitCommand(repoDir, "add", oldFile); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}
	if err := runGitCommand(repoDir, "commit", "-m", "Add file to rename"); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Rename and stage
	if err := runGitCommand(repoDir, "mv", oldFile, newFile); err != nil {
		t.Fatalf("Failed to rename file: %v", err)
	}

	// Test backup
	backupFile := filepath.Join(repoDir, "backup.txt")
	err := runInDir(repoDir, func() error {
		return HandleBackup(backupFile)
	})

	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify backup content
	content, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	backupStr := string(content)
	if !strings.Contains(backupStr, "===== old-name.txt -> new-name.txt =====") {
		t.Error("Backup missing renamed file header")
	}
	if !strings.Contains(backupStr, "content to be renamed") {
		t.Error("Backup missing renamed file content")
	}
}

func testRestoreCleanWorktree(t *testing.T, repoDir string) {
	// Create backup file outside the repo to avoid making worktree dirty
	backupContent := `===== test-file.txt =====
test content line 1
test content line 2

===== subdir/another.go =====
package main

func test() {
	println("test")
}

`
	backupFile := filepath.Join(os.TempDir(), "backup-clean-test.txt")
	if err := os.WriteFile(backupFile, []byte(backupContent), 0644); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}
	defer os.Remove(backupFile)

	// Test restore
	err := runInDir(repoDir, func() error {
		return HandleRestore(backupFile)
	})

	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify files were created and staged
	if _, err := os.Stat(filepath.Join(repoDir, "test-file.txt")); os.IsNotExist(err) {
		t.Error("test-file.txt was not created")
	}
	if _, err := os.Stat(filepath.Join(repoDir, "subdir", "another.go")); os.IsNotExist(err) {
		t.Error("subdir/another.go was not created")
	}

	// Check if files are staged
	output, err := runGitCommandOutput(repoDir, "diff", "--cached", "--name-only")
	if err != nil {
		t.Fatalf("Failed to check staged files: %v", err)
	}

	stagedFiles := strings.TrimSpace(output)
	if !strings.Contains(stagedFiles, "test-file.txt") {
		t.Error("test-file.txt is not staged")
	}
	if !strings.Contains(stagedFiles, "subdir/another.go") {
		t.Error("subdir/another.go is not staged")
	}
}

func testRestoreDirtyWorktree(t *testing.T, repoDir string) {
	// Create a file to make worktree dirty
	dirtyFile := filepath.Join(repoDir, "dirty.txt")
	if err := os.WriteFile(dirtyFile, []byte("dirty content"), 0644); err != nil {
		t.Fatalf("Failed to create dirty file: %v", err)
	}

	// Create backup file
	backupContent := `===== test-file.txt =====
test content
`
	backupFile := filepath.Join(repoDir, "backup.txt")
	if err := os.WriteFile(backupFile, []byte(backupContent), 0644); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}

	// Test restore should fail
	err := runInDir(repoDir, func() error {
		return HandleRestore(backupFile)
	})

	if err == nil {
		t.Error("Expected error for dirty worktree, got nil")
	}
	if !strings.Contains(err.Error(), "worktree is not clean") {
		t.Errorf("Expected 'worktree is not clean' error, got: %v", err)
	}
}

func testRestoreInvalidBackup(t *testing.T, repoDir string) {
	// Create invalid backup file outside the repo
	backupContent := `invalid backup format
no headers here
`
	backupFile := filepath.Join(os.TempDir(), "invalid-backup-test.txt")
	if err := os.WriteFile(backupFile, []byte(backupContent), 0644); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}
	defer os.Remove(backupFile)

	// Test restore
	err := runInDir(repoDir, func() error {
		return HandleRestore(backupFile)
	})

	if err == nil {
		t.Error("Expected error for invalid backup, got nil")
	}
	if !strings.Contains(err.Error(), "no files found in backup") {
		t.Errorf("Expected 'no files found in backup' error, got: %v", err)
	}
}

func testBackupRestoreRoundTrip(t *testing.T, repoDir string) {
	// Create various types of files
	files := map[string]string{
		"regular.txt":       "regular file content\nwith multiple lines",
		"subdir/nested.go":  "package main\n\nfunc main() {\n\tprintln(\"nested\")\n}",
		"special-chars.txt": "content with special chars: !@#$%^&*()",
		"empty.txt":         "",
		"unicode.txt":       "Unicode content: 你好世界 🌍",
	}

	// Create and stage files
	for relPath, content := range files {
		fullPath := filepath.Join(repoDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", relPath, err)
		}
	}

	// Stage all files
	if err := runGitCommand(repoDir, "add", "."); err != nil {
		t.Fatalf("Failed to stage files: %v", err)
	}

	// Backup to external file
	backupFile := filepath.Join(os.TempDir(), "roundtrip-backup-test.txt")
	err := runInDir(repoDir, func() error {
		return HandleBackup(backupFile)
	})
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}
	defer os.Remove(backupFile)

	// Reset to clean state
	if err := runGitCommand(repoDir, "reset", "--hard"); err != nil {
		t.Fatalf("Failed to reset: %v", err)
	}
	if err := runGitCommand(repoDir, "clean", "-fd"); err != nil {
		t.Fatalf("Failed to clean: %v", err)
	}

	// Restore
	err = runInDir(repoDir, func() error {
		return HandleRestore(backupFile)
	})
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify all files are restored with correct content
	for relPath, expectedContent := range files {
		fullPath := filepath.Join(repoDir, relPath)
		actualContent, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("Failed to read restored file %s: %v", relPath, err)
			continue
		}
		if string(actualContent) != expectedContent {
			t.Errorf("Content mismatch for %s:\nExpected: %q\nActual: %q", relPath, expectedContent, string(actualContent))
		}
	}

	// Verify all files are staged
	output, err := runGitCommandOutput(repoDir, "diff", "--cached", "--name-only")
	if err != nil {
		t.Fatalf("Failed to check staged files: %v", err)
	}

	stagedFiles := strings.Split(strings.TrimSpace(output), "\n")
	if len(stagedFiles) != len(files) {
		t.Errorf("Expected %d staged files, got %d", len(files), len(stagedFiles))
	}
}

// Helper functions

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

func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

func runGitCommandOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	return string(output), err
}
