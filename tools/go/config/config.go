package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/kool/tools/config"
)

// LocalModulesConfig represents the configuration for local modules operations
type LocalModulesConfig struct {
	LocalModules []string `json:"local_module_paths"`
}

// ModuleUpdateInfo represents information about a module that needs updating
type ModuleUpdateInfo struct {
	ModulePath     string // The module path (e.g., github.com/user/repo)
	LocalPath      string // Local filesystem path
	CurrentVersion string // Current version in go.mod
	LatestVersion  string // Latest clean version (without prefix)
	LatestTag      string // Latest tag to update to
	IsReplacement  bool   // Whether this is currently a replacement
}

// GetLocalModulesConfig reads the local modules configuration from the config file
func GetLocalModulesConfig() (*LocalModulesConfig, error) {
	koolConfigDir, err := config.GetKoolConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get kool config directory: %w", err)
	}

	configPath := filepath.Join(koolConfigDir, "go_update_all.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s (run 'kool go replace --all --show' or 'kool go update --all --show' first)", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg LocalModulesConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// ShowLocalModulesConfig prints the path to go_update_all.json and creates the file and its directory if not exists yet
func ShowLocalModulesConfig() error {
	koolConfigDir, err := config.GetKoolConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get kool config directory: %w", err)
	}

	configPath := filepath.Join(koolConfigDir, "go_update_all.json")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(koolConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		defaultConfig := LocalModulesConfig{
			LocalModules: []string{},
		}

		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal default config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	fmt.Println(configPath)
	return nil
}
