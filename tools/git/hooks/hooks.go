package hooks

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed git-hooks/main.go
var gitHooksMainGo string

func HandleInit(args []string) error {
	targetDir := filepath.Join("script", "git-hooks")
	targetFile := filepath.Join(targetDir, "main.go")

	// Check if target file already exists
	if _, err := os.Stat(targetFile); err == nil {
		return fmt.Errorf("file %s already exists", targetFile)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Write the embedded file to the target location
	if err := os.WriteFile(targetFile, []byte(gitHooksMainGo), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", targetFile, err)
	}

	fmt.Printf("Git hooks initialized successfully, you can run: \n")
	fmt.Printf("  go run ./script/git-hooks install\n")

	return nil
}
