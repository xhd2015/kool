package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMainFunctions(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "rules_main_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test rule file
	testRulePath := filepath.Join(tempDir, "test_rule.mdc")
	testContent := []byte("test rule content")
	if err := os.WriteFile(testRulePath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test rule file: %v", err)
	}

	// Test Add, List, and Use functions in a real integration-like test
	t.Run("IntegrationTest", func(t *testing.T) {
		// First, check the GetRulesDir function
		rulesDir, err := GetRulesDir()
		if err != nil {
			t.Fatalf("GetRulesDir() failed: %v", err)
		}

		// Check if the directory exists
		if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
			t.Errorf("GetRulesDir() did not create the directory: %s", rulesDir)
		}

		// Check the GetCursorRulesDir function
		cursorRulesDir, err := GetCursorRulesDir()
		if err != nil {
			t.Fatalf("GetCursorRulesDir() failed: %v", err)
		}

		// Check if the directory exists
		if _, err := os.Stat(cursorRulesDir); os.IsNotExist(err) {
			t.Errorf("GetCursorRulesDir() did not create the directory: %s", cursorRulesDir)
		}

		// Note: We can't easily test the actual Add, List, Use functions
		// without affecting the user's actual ~/.kool/rules directory
		// or the current project's .cursor/rules directory.
		// A proper integration test would require mocking file systems or specific
		// environment/directory preparation.
	})

	// Test the FileExists function directly
	t.Run("FileExists", func(t *testing.T) {
		// Test with existing file
		if !fileExists(testRulePath) {
			t.Errorf("fileExists() = false for existing file: %s", testRulePath)
		}

		// Test with non-existent file
		nonExistentPath := testRulePath + "_nonexistent"
		if fileExists(nonExistentPath) {
			t.Errorf("fileExists() = true for non-existent file: %s", nonExistentPath)
		}
	})
}
