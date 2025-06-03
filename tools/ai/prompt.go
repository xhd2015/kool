package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func HandlePrompt(dirPath string) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("error resolving path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid directory: %s", absPath)
	}

	// Collect directory structure and file contents
	output, err := collectDirStructureAndFiles(absPath)
	if err != nil {
		return fmt.Errorf("error collecting directory structure and files: %w", err)
	}

	// Print the result
	fmt.Println(output)
	return nil
}

// collectDirStructureAndFiles collects the directory structure and file contents.
func collectDirStructureAndFiles(root string) (string, error) {
	var builder strings.Builder

	// Write directory structure header
	builder.WriteString("Directory Structure:\n")
	builder.WriteString("===================\n")

	// Walk the directory to collect structure first
	var dirStructure []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Calculate relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if relPath == "." {
			return nil
		}
		// Add indentation based on path depth
		depth := len(strings.Split(relPath, string(os.PathSeparator)))
		indent := strings.Repeat("  ", depth-1)
		if info.IsDir() {
			dirStructure = append(dirStructure, fmt.Sprintf("%s📁 %s", indent, relPath))
		} else {
			dirStructure = append(dirStructure, fmt.Sprintf("%s📄 %s", indent, relPath))
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error walking directory: %w", err)
	}

	// Write the directory structure
	for _, line := range dirStructure {
		builder.WriteString(line + "\n")
	}

	// Write file contents header
	builder.WriteString("\nFile Contents:\n")
	builder.WriteString("===================\n")

	// Walk again to collect file contents
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories and the root itself
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		// Write file header and content
		builder.WriteString(fmt.Sprintf("\nFile: %s\n", relPath))
		builder.WriteString("-----\n")
		builder.Write(content)
		if !strings.HasSuffix(string(content), "\n") {
			builder.WriteString("\n")
		}
		builder.WriteString("-----\n")

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error reading files: %w", err)
	}

	return builder.String(), nil
}
