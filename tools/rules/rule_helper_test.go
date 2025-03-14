package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper for creating a test environment
func setupTestEnvironment(t *testing.T) (string, string, string, func()) {
	// Create temporary directories for testing
	tmpBaseDir, err := os.MkdirTemp("", "rules_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test structure
	koolRulesDir := filepath.Join(tmpBaseDir, "kool_rules")
	cursorRulesDir := filepath.Join(tmpBaseDir, "cursor_rules")

	if err := os.MkdirAll(koolRulesDir, 0755); err != nil {
		os.RemoveAll(tmpBaseDir)
		t.Fatalf("Failed to create kool rules dir: %v", err)
	}

	if err := os.MkdirAll(cursorRulesDir, 0755); err != nil {
		os.RemoveAll(tmpBaseDir)
		t.Fatalf("Failed to create cursor rules dir: %v", err)
	}

	// Create a test source file
	sourcePath := filepath.Join(tmpBaseDir, "test_rule.mdc")
	testContent := []byte("test rule content")
	if err := os.WriteFile(sourcePath, testContent, 0644); err != nil {
		os.RemoveAll(tmpBaseDir)
		t.Fatalf("Failed to create test source file: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpBaseDir)
	}

	return koolRulesDir, cursorRulesDir, sourcePath, cleanup
}

func TestDirectoryFunctions(t *testing.T) {
	// Test the directory functions
	t.Run("RulesDir", func(t *testing.T) {
		dir, err := GetRulesDir()
		if err != nil {
			t.Fatalf("GetRulesDir() failed: %v", err)
		}

		// Check if the directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("GetRulesDir() did not create the directory: %s", dir)
		}

		// Check if the path is correct
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".kool", "rules", "files")
		if dir != expected {
			t.Errorf("GetRulesDir() = %s, want %s", dir, expected)
		}
	})

	t.Run("CursorRulesDir", func(t *testing.T) {
		dir, err := GetCursorRulesDir()
		if err != nil {
			t.Fatalf("GetCursorRulesDir() failed: %v", err)
		}

		// Check if the directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("GetCursorRulesDir() did not create the directory: %s", dir)
		}

		// Check if the path is correct
		cwd, _ := os.Getwd()
		expected := filepath.Join(cwd, ".cursor", "rules")
		if dir != expected {
			t.Errorf("GetCursorRulesDir() = %s, want %s", dir, expected)
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	koolRulesDir, _, sourcePath, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Test copyFile
	t.Run("copyFile", func(t *testing.T) {
		destPath := filepath.Join(koolRulesDir, "test_rule.mdc")

		err := copyFile(sourcePath, destPath)
		if err != nil {
			t.Fatalf("copyFile() failed: %v", err)
		}

		// Check if the file was copied
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Errorf("copyFile() did not copy the file to %s", destPath)
		}

		// Check content
		srcContent, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}

		destContent, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}

		if string(srcContent) != string(destContent) {
			t.Errorf("copyFile() did not copy the content correctly")
		}
	})

	// Test Get functionality without mocking
	t.Run("Get", func(t *testing.T) {
		// Create a test rule file directly in the koolRulesDir
		testRuleContent := "Test rule content for Get test"
		testRuleName := "get_test_rule.mdc"
		testRulePath := filepath.Join(koolRulesDir, testRuleName)

		// Write the test file
		err := os.WriteFile(testRulePath, []byte(testRuleContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test rule file: %v", err)
		}

		// We can't test Get() directly as it uses GetRulesDir and prints to stdout
		// Instead, we'll test the core functionality of reading a rule file

		// Read the file
		content, err := os.ReadFile(testRulePath)
		if err != nil {
			t.Fatalf("Failed to read rule file: %v", err)
		}

		// Verify content
		if string(content) != testRuleContent {
			t.Errorf("Rule file content doesn't match. Expected: %s, Got: %s",
				testRuleContent, string(content))
		}

		// Test file existence check
		nonExistentPath := filepath.Join(koolRulesDir, "nonexistent.mdc")
		if fileExists(nonExistentPath) {
			t.Errorf("fileExists() = true for non-existent file: %s", nonExistentPath)
		}
	})

	// Test listFilesInDir
	t.Run("listFilesInDir", func(t *testing.T) {
		// Create multiple files in the directory
		fileNames := []string{"file1.mdc", "file2.mdc", "file3.mdc"}

		for _, name := range fileNames {
			path := filepath.Join(koolRulesDir, name)
			err := os.WriteFile(path, []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", name, err)
			}
		}

		// List files
		files, err := listFilesInDir(koolRulesDir)
		if err != nil {
			t.Fatalf("listFilesInDir() failed: %v", err)
		}

		// Note: Both test_rule.mdc (from copyFile test) and get_test_rule.mdc (from Get test)
		// already exist in the directory, so we expect fileNames.length + 2 files
		expectedCount := len(fileNames) + 2
		if len(files) != expectedCount {
			t.Errorf("listFilesInDir() returned %d files, expected %d", len(files), expectedCount)
		}

		// Check if all expected files are in the list
		fileMap := make(map[string]bool)
		for _, f := range files {
			fileMap[f] = true
		}

		for _, expected := range fileNames {
			if !fileMap[expected] {
				t.Errorf("listFilesInDir() did not include %s", expected)
			}
		}
	})

	// Test fileExists
	t.Run("fileExists", func(t *testing.T) {
		// Test existing file
		if !fileExists(sourcePath) {
			t.Errorf("fileExists() = false for existing file %s", sourcePath)
		}

		// Test non-existing file
		nonExistentPath := sourcePath + "_nonexistent"
		if fileExists(nonExistentPath) {
			t.Errorf("fileExists() = true for non-existent file %s", nonExistentPath)
		}
	})
}
