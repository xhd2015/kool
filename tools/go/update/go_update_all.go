package update

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/kool/tools/git/tag"
	"github.com/xhd2015/kool/tools/go/resolve"
)

type replacementInfo struct {
	modulePath string
	targetDir  string
	tag        string
	tagPrefix  string
}

func UpdateAll() error {
	// Get current directory's go.mod info
	modInfo, err := resolve.GetModuleInfo(".")
	if err != nil {
		return fmt.Errorf("failed to get module info: %w", err)
	}

	if len(modInfo.Replace) == 0 {
		fmt.Fprintf(os.Stderr, "no replacements found\n")
		return nil
	}

	// Step 1: Collect and validate all local filesystem replacements
	var replacements []replacementInfo
	for _, repl := range modInfo.Replace {
		// Skip non-local replacements (those with versions)
		if repl.New.Version != "" {
			continue
		}

		targetDir := repl.New.Path
		if !filepath.IsAbs(targetDir) {
			targetDir, err = filepath.Abs(targetDir)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for %s: %w", repl.New.Path, err)
			}
		}

		// Check if target directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			return fmt.Errorf("replacement target directory does not exist: %s", targetDir)
		}

		// Get target directory's module info
		targetModInfo, err := resolve.GetModuleInfo(targetDir)
		if err != nil {
			return fmt.Errorf("failed to get module info for %s: %w", targetDir, err)
		}
		oldModPath := repl.Old.Path
		targetModPath := targetModInfo.Module.Path

		// Step 1: Check if module path matches exactly or is a submodule
		if oldModPath != targetModPath {
			fmt.Fprintf(os.Stderr, "skipping replacement %s => %s: module path mismatch (target module: %s)\n",
				oldModPath, targetDir, targetModPath)
			continue
		}

		// Get the root module path (the main module in the repo)
		rootModPath, err := resolve.GetRootModulePath(targetDir)
		if err != nil {
			return err
		}

		// Calculate tag prefix for nested submodules
		subModulePath, ok := cutSubmoduleSuffix(rootModPath, targetModPath)
		if !ok {
			fmt.Fprintf(os.Stderr, "skipping replacement %s => %s: module path is not a submodule of the root module\n",
				oldModPath, targetDir)
			continue
		}
		var tagPrefix string
		if subModulePath != "" {
			tagPrefix = subModulePath + "/"
		}

		versionTag, err := tag.GetVersionTagAtHEAD(targetDir, tagPrefix)
		if err != nil {
			return err
		}

		replacements = append(replacements, replacementInfo{
			modulePath: repl.Old.Path,
			targetDir:  targetDir,
			tag:        versionTag,
			tagPrefix:  tagPrefix,
		})
	}

	if len(replacements) == 0 {
		fmt.Fprintf(os.Stderr, "no local replacements to update\n")
		return nil
	}

	// Step 3: All validations passed, now update all replacements
	for _, repl := range replacements {
		fmt.Fprintf(os.Stderr, "updating %s to %s\n", repl.modulePath, repl.tag)

		// Drop the replacement first, then update the version
		version := strings.TrimPrefix(repl.tag, repl.tagPrefix)
		if err := GoModDropReplace(repl.modulePath); err != nil {
			return fmt.Errorf("failed to drop replacement for %s: %w", repl.modulePath, err)
		}
		if err := GoModEditRequire(repl.modulePath, version); err != nil {
			return fmt.Errorf("failed to update %s: %w", repl.modulePath, err)
		}
	}

	fmt.Fprintf(os.Stderr, "successfully updated %d replacement(s)\n", len(replacements))
	return nil
}

// calculateTagPrefix calculates the tag prefix for a given module path
// For nested submodules, returns "path/to/submodule/"
// For root modules, returns ""
func calculateTagPrefix(targetDir, modulePath string) (string, error) {
	// Get the root module path (the main module in the repo)
	rootModPath, err := resolve.GetRootModulePath(targetDir)
	if err != nil {
		return "", err
	}

	// If this is a nested submodule, extract the submodule path
	if modulePath != rootModPath {
		if !strings.HasPrefix(modulePath, rootModPath+"/") {
			return "", fmt.Errorf("module path %s does not start with root module path %s", modulePath, rootModPath)
		}
		subModulePath := strings.TrimPrefix(modulePath, rootModPath+"/")
		return subModulePath + "/", nil
	}

	// Root module has no prefix
	return "", nil
}

func cutSubmoduleSuffix(parentModulePath, childModulePath string) (string, bool) {
	if !strings.HasPrefix(childModulePath, parentModulePath) {
		return "", false
	}
	if len(childModulePath) == len(parentModulePath) {
		return "", true
	}
	if childModulePath[len(parentModulePath)] != '/' {
		return "", false
	}
	return childModulePath[len(parentModulePath)+1:], true
}
