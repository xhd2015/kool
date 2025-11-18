package resolve

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/git"
)

type ModuleInfo struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Require []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Require"`
	Replace []struct {
		Old struct {
			Path string `json:"Path"`
		} `json:"Old"`
		New struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"New"`
	} `json:"Replace"`
}

func GetModuleInfo(dir string) (*ModuleInfo, error) {
	// Run 'go mod edit -json' in the directory
	output, err := cmd.Dir(dir).Output("go", "mod", "edit", "-json")
	if err != nil {
		return nil, fmt.Errorf("resolve go mod: %s %w", dir, err)
	}

	// Parse the JSON output
	var modInfo ModuleInfo
	if err := json.Unmarshal([]byte(output), &modInfo); err != nil {
		return nil, fmt.Errorf("failed to parse module info: %v", err)
	}

	return &modInfo, nil
}

// GetRootModulePath gets the module path from the go.mod at the root of the git repository
func GetRootModulePath(targetDir string) (string, error) {
	// Find the git root
	gitRoot, err := git.ShowTopLevel(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to get git root for %s: %w", targetDir, err)
	}

	// Get the module info from the git root
	rootModInfo, err := GetModuleInfo(gitRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get root module info for %s: %w", gitRoot, err)
	}

	return rootModInfo.Module.Path, nil
}

// LocalModuleInfo represents information about a resolved local module
type LocalModuleInfo struct {
	LocalPath      string      // Local filesystem path (absolute)
	ModuleInfo     *ModuleInfo // Module information from go.mod
	IsDependency   bool        // Whether this module is a dependency of the current module
	IsReplaced     bool        // Whether this module is currently replaced
	CurrentVersion string      // Current version in go.mod (if it's a dependency)
}

// ResolveLocalModules resolves a list of local module directories and checks their dependency status
// against the current directory's go.mod file.
func ResolveLocalModules(currentDir string, localModDirs []string) (*ModuleInfo, []*LocalModuleInfo, error) {
	// Get current directory's go.mod info
	currentModInfo, err := GetModuleInfo(currentDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current module info: %w", err)
	}

	var resolvedModules []*LocalModuleInfo

	for _, localModDir := range localModDirs {
		resolved, err := resolveLocalModule(localModDir, currentModInfo)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve local module %s: %w", localModDir, err)
		}
		if resolved != nil {
			resolvedModules = append(resolvedModules, resolved)
		}
	}

	return currentModInfo, resolvedModules, nil
}

// resolveLocalModule resolves a single local module directory
func resolveLocalModule(localModDir string, currentModInfo *ModuleInfo) (*LocalModuleInfo, error) {
	// Get absolute path
	absPath, err := filepath.Abs(localModDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", localModDir, err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("local module directory does not exist: %s: %w", absPath, err)
	}

	// Get module info from the local directory
	localModInfo, err := GetModuleInfo(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get module info for %s: %w", absPath, err)
	}

	if localModInfo.Module.Path == "" {
		return nil, fmt.Errorf("not a go module: %s", absPath)
	}

	modulePath := localModInfo.Module.Path

	// Check dependency status
	var isDependency bool
	var isReplaced bool
	var currentVersion string

	// Check in requirements
	for _, req := range currentModInfo.Require {
		if req.Path == modulePath {
			isDependency = true
			currentVersion = req.Version
			break
		}
	}

	// Check in replacements
	for _, repl := range currentModInfo.Replace {
		if repl.Old.Path == modulePath {
			isReplaced = true
			if !isDependency {
				// If it's only in replacements, still consider it as a dependency
				isDependency = true
			}
			break
		}
	}

	return &LocalModuleInfo{
		LocalPath:      absPath,
		ModuleInfo:     localModInfo,
		IsDependency:   isDependency,
		IsReplaced:     isReplaced,
		CurrentVersion: currentVersion,
	}, nil
}
