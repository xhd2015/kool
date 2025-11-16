package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetRulesDir returns the path to the .kool/rules directory in the user's home directory
func GetKoolConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".kool"), nil
}
