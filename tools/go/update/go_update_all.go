package update

import (
	"fmt"
	"os"
	"path/filepath"

	goconfig "github.com/xhd2015/kool/tools/go/config"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	gotoolupdate "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
)

// ModuleUpdateInfo represents information about a module that needs updating
type ModuleUpdateInfo struct {
	ModulePath     string // The module path (e.g., github.com/user/repo)
	LocalPath      string // Local filesystem path
	CurrentVersion string // Current version in go.mod
	LatestVersion  string // Latest clean version (without prefix)
	LatestTag      string // Latest tag to update to
	IsReplacement  bool   // Whether this is currently a replacement
}

// UpdateAll reads the config, gets the LocalModules list, and updates dependencies to latest local tags
func UpdateAll(dir string) error {
	config, err := goconfig.GetLocalModulesConfig()
	if err != nil {
		return err
	}

	if len(config.LocalModules) == 0 {
		fmt.Printf("No local modules configured\n")
		return nil
	}

	// Get current directory's go.mod info
	currentModInfo, err := resolve.GetModuleInfo(dir)
	if err != nil {
		return fmt.Errorf("failed to get current module info: %w", err)
	}

	// Collect from existing replacements
	replaceInfos, err := collectReplacementUpdateInfos(currentModInfo)
	if err != nil {
		return fmt.Errorf("failed to collect replacement info: %w", err)
	}

	// Phase 1: Check and collect all module update info
	var updateInfos []ModuleUpdateInfo

	// Resolve all local modules and their dependency status
	_, resolvedModules, err := resolve.ResolveLocalModules(dir, config.LocalModules)
	if err != nil {
		return err
	}

	// Collect update info from resolved modules
	for _, resolved := range resolvedModules {
		if !resolved.IsDependency {
			// fmt.Printf("Skipping module %s: no dependency found\n", resolved.ModuleInfo.Module.Path)
			continue
		}

		info, err := buildModuleUpdateInfo(resolved)
		if err != nil {
			fmt.Printf("Skipping module %s: %v\n", resolved.ModuleInfo.Module.Path, err)
			continue
		}
		if info != nil {
			updateInfos = append(updateInfos, *info)
		}
	}

	// Merge replacement infos, avoiding duplicates
	updateInfos = mergeUpdateInfos(updateInfos, replaceInfos)

	if len(updateInfos) == 0 {
		fmt.Printf("No modules to update\n")
		return nil
	}

	// Phase 2: Execute updates
	return executeModuleUpdates(dir, updateInfos)
}

// buildModuleUpdateInfo builds update information from a resolved local module
func buildModuleUpdateInfo(resolved *resolve.LocalModuleInfo) (*ModuleUpdateInfo, error) {
	modulePath := resolved.ModuleInfo.Module.Path
	absPath := resolved.LocalPath

	versionPrefix, err := gotoolupdate.CalculateVersionPrefix(absPath, modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate version prefix for %s: %w", modulePath, err)
	}

	latestTag, err := gotoolupdate.GetLatestVersionTag(absPath, versionPrefix)
	if err != nil {
		return nil, fmt.Errorf("no suitable tag found for %s: %w", modulePath, err)
	}

	cleanLatestVersion := gotoolupdate.StripVersionPrefix(versionPrefix, latestTag)
	if !isValidVersionTag(cleanLatestVersion) {
		fmt.Fprintf(os.Stderr, "  Latest version %s is not a valid semantic version, skipping\n", cleanLatestVersion)
		return nil, nil
	}

	currentVersion := resolved.CurrentVersion
	cleanCurrentVersion := gotoolupdate.StripVersionPrefix(versionPrefix, currentVersion)
	if currentVersion != "" && !isValidVersionTag(cleanCurrentVersion) {
		fmt.Fprintf(os.Stderr, "  Current version %s is not a valid semantic version, skipping\n", cleanCurrentVersion)
		return nil, nil
	}

	// Check if we need to update - only update if local version is newer
	if currentVersion != "" && !isNewerVersion(cleanLatestVersion, cleanCurrentVersion) {
		fmt.Fprintf(os.Stderr, "  Local version %s is not newer than current %s, skipping\n", cleanLatestVersion, cleanCurrentVersion)
		return nil, nil
	}

	return &ModuleUpdateInfo{
		ModulePath:     modulePath,
		LocalPath:      absPath,
		CurrentVersion: currentVersion,
		LatestTag:      latestTag,
		LatestVersion:  cleanLatestVersion,
		IsReplacement:  resolved.IsReplaced,
	}, nil
}

// collectReplacementUpdateInfos collects update information from existing go.mod replacements
func collectReplacementUpdateInfos(currentModInfo *resolve.ModuleInfo) ([]ModuleUpdateInfo, error) {
	var infos []ModuleUpdateInfo

	for _, repl := range currentModInfo.Replace {
		// Skip non-local replacements (those with versions)
		if repl.New.Version != "" {
			continue
		}

		targetDir := repl.New.Path
		if !filepath.IsAbs(targetDir) {
			absPath, err := filepath.Abs(targetDir)
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path for %s: %w", repl.New.Path, err)
			}
			targetDir = absPath
		}

		// Check if target directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			fmt.Printf("  Replacement target directory does not exist: %s, skipping\n", targetDir)
			continue
		}

		// Get target directory's module info
		targetModInfo, err := resolve.GetModuleInfo(targetDir)
		if err != nil {
			fmt.Printf("  Failed to get module info for %s: %v, skipping\n", targetDir, err)
			continue
		}

		oldModPath := repl.Old.Path
		targetModPath := targetModInfo.Module.Path

		// Check if module path matches
		if oldModPath != targetModPath {
			fmt.Printf("  Skipping replacement %s => %s: module path mismatch (target module: %s)\n",
				oldModPath, targetDir, targetModPath)
			continue
		}

		versionPrefix, err := gotoolupdate.CalculateVersionPrefix(targetDir, targetModPath)
		if err != nil {
			fmt.Printf("  Failed to calculate version prefix for %s: %v, skipping\n", targetModPath, err)
			continue
		}

		latestTag, err := gotoolupdate.GetLatestVersionTag(targetDir, versionPrefix)
		if err != nil {
			fmt.Printf("  No suitable tag found for %s: %v, skipping\n", targetModPath, err)
			continue
		}

		cleanLatestVersion := gotoolupdate.StripVersionPrefix(versionPrefix, latestTag)
		if !isValidVersionTag(cleanLatestVersion) {
			return nil, fmt.Errorf("replacement %s has invalid version %s: not a valid semantic version", targetModPath, cleanLatestVersion)
		}

		// Find the current version in go.mod requirements
		var currentVersion string
		for _, req := range currentModInfo.Require {
			if req.Path == oldModPath {
				currentVersion = req.Version
				break
			}
		}

		infos = append(infos, ModuleUpdateInfo{
			ModulePath:     oldModPath,
			LocalPath:      targetDir,
			CurrentVersion: currentVersion,
			LatestTag:      latestTag,
			LatestVersion:  cleanLatestVersion,
			IsReplacement:  true,
		})
	}

	return infos, nil
}

// mergeUpdateInfos merges two slices of ModuleUpdateInfo, avoiding duplicates
func mergeUpdateInfos(updates, replacements []ModuleUpdateInfo) []ModuleUpdateInfo {
	merged := make([]ModuleUpdateInfo, 0, len(updates)+len(replacements))
	// Create a map to track existing modules
	existingMap := make(map[string]bool)
	for _, info := range replacements {
		existingMap[info.ModulePath] = true
		merged = append(merged, info)
	}

	// Add non-duplicate additional infos
	for _, info := range updates {
		if !existingMap[info.ModulePath] {
			merged = append(merged, info)
		}
	}

	return merged
}

// executeModuleUpdates executes the actual updates for all collected module infos
func executeModuleUpdates(dir string, updateInfos []ModuleUpdateInfo) error {
	fmt.Printf("Updating %d module(s):\n", len(updateInfos))

	for _, info := range updateInfos {
		currentVersion := info.CurrentVersion
		latestVersion := info.LatestVersion
		targetModPath := info.ModulePath

		// For replacements: drop replacement first
		if info.IsReplacement {
			if err := commands.GoModDropReplace(info.ModulePath, &commands.GoModEditOptions{Dir: dir}); err != nil {
				return fmt.Errorf("failed to drop replacement for %s: %w", info.ModulePath, err)
			}
		}

		// If we have a real current version, validate and compare
		if currentVersion != "" && isValidVersionTag(currentVersion) {
			// Only update if local version is newer
			if !isNewerVersion(latestVersion, currentVersion) {
				fmt.Printf("  %s: local version %s is not newer than current %s, skipping updating\n", targetModPath, latestVersion, currentVersion)
				continue
			}
		}
		// Update the module version (same for both replacements and regular dependencies)
		if err := commands.GoModEditRequire(targetModPath, latestVersion, &commands.GoModEditOptions{
			Dir: dir,
		}); err != nil {
			return fmt.Errorf("failed to update %s: %w", targetModPath, err)
		}

		fmt.Printf("  Successfully updated %s to %s\n", targetModPath, latestVersion)
	}

	return nil
}
