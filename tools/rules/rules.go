package rules

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// GetRulesDir returns the path to the .kool/rules directory in the user's home directory
func GetRulesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	dir := filepath.Join(home, ".kool", "rules", "files")

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create rules directory: %w", err)
	}

	return dir, nil
}

// GetCursorRulesDir returns the path to the .cursor/rules directory in the current working directory
func GetCursorRulesDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	dir := filepath.Join(cwd, ".cursor", "rules")

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .cursor/rules directory: %w", err)
	}

	return dir, nil
}

// Add copies a rule file to the ~/.kool/rules/files directory
// If a file with the same name already exists, a timestamp will be appended to make it unique
func Add(rulePath string) (string, error) {
	if !fileExists(rulePath) {
		return "", fmt.Errorf("file not found: %s", rulePath)
	}

	rulesDir, err := GetRulesDir()
	if err != nil {
		return "", err
	}

	fileName := filepath.Base(rulePath)
	destPath := filepath.Join(rulesDir, fileName)
	resultName := fileName

	// If file already exists, append a date-time suffix
	if fileExists(destPath) {
		// Generate timestamp suffix in format YYYYMMDD-HHMMSS
		timestamp := time.Now().Format("20060102-150405")

		// Split filename and extension to insert timestamp
		ext := filepath.Ext(fileName)
		baseName := fileName[:len(fileName)-len(ext)]

		// Create new filename with timestamp
		resultName = fmt.Sprintf("%s-%s%s", baseName, timestamp, ext)
		destPath = filepath.Join(rulesDir, resultName)
	}

	err = copyFile(rulePath, destPath)
	return resultName, err
}

// List returns a list of rule files in the ~/.kool/rules/files directory
func List() ([]string, error) {
	rulesDir, err := GetRulesDir()
	if err != nil {
		return nil, err
	}

	return listFilesInDir(rulesDir)
}

// Use copies a rule file from ~/.kool/rules/files to .cursor/rules if it doesn't already exist
func Use(ruleName string) error {
	cursorRulesDir, err := GetCursorRulesDir()
	if err != nil {
		return err
	}

	cursorRulePath := filepath.Join(cursorRulesDir, ruleName)

	// Check if file already exists in .cursor/rules
	if fileExists(cursorRulePath) {
		return fmt.Errorf("rule already exists in .cursor/rules: %s", ruleName)
	}

	// Find the rule in ~/.kool/rules/files
	rulesDir, err := GetRulesDir()
	if err != nil {
		return err
	}

	rulePath := filepath.Join(rulesDir, ruleName)
	if !fileExists(rulePath) {
		return fmt.Errorf("rule not found in ~/.kool/rules/files: %s", ruleName)
	}

	// Copy the rule to .cursor/rules
	return copyFile(rulePath, cursorRulePath)
}

// Remove deletes a rule file from the ~/.kool/rules/files directory
func Remove(ruleName string) error {
	rulesDir, err := GetRulesDir()
	if err != nil {
		return err
	}

	rulePath := filepath.Join(rulesDir, ruleName)
	if !fileExists(rulePath) {
		return fmt.Errorf("rule not found: %s", ruleName)
	}

	return os.Remove(rulePath)
}

// Helper functions

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// listFilesInDir returns a list of files in a directory
func listFilesInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
