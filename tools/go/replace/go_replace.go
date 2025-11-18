package replace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/kool/tools/go/commands"
	goconfig "github.com/xhd2015/kool/tools/go/config"
	"github.com/xhd2015/kool/tools/go/resolve"
)

// equivalent to go mod edit -replace "$mod=$absDir":
//
// dir=$1
// if [[ -z $dir ]];then
//     echo "requires dir" >&2
//     exit 1
// fi

// if [[ ! -d $dir ]];then
//     echo "no such dir: $dir" >&2
//     exit 1
// fi

// mod=$(cd "$dir" && go mod edit -json|jq -r '.Module.Path')
// if [[ -z $mod ]];then
//     echo "not a go module: $dir" >&2
//     exit 1
// fi
// absDir=$(cd "$dir" && pwd)
// go mod edit -replace "$mod=$absDir"

// Replace checks if the given directory exists and contains a valid Go module,
// returns the absolute path of the directory and the module path
func Replace(dir string) (absDir string, modulePath string, err error) {
	// Check if directory is provided
	if dir == "" {
		return "", "", fmt.Errorf("requires dir")
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", "", fmt.Errorf("no such dir: %s", dir)
	}

	// Get absolute path
	absDir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	modInfo, err := resolve.GetModuleInfo(dir)
	if err != nil {
		return "", "", err
	}

	if modInfo.Module.Path == "" {
		return "", "", fmt.Errorf("not a go module: %s", dir)
	}
	// executing go mod edit -replace "$mod=$absDir"
	err = commands.GoModEditReplace(modInfo.Module.Path, absDir, nil)
	if err != nil {
		return "", "", err
	}

	return absDir, modInfo.Module.Path, nil
}

// ReplaceAll reads the config, gets the LocalModules list, and adds replace directives for dependencies
func ReplaceAll() error {
	config, err := goconfig.GetLocalModulesConfig()
	if err != nil {
		return err
	}

	if len(config.LocalModules) == 0 {
		fmt.Printf("No local modules configured\n")
		return nil
	}

	// Resolve all local modules and their dependency status
	_, resolvedModules, err := resolve.ResolveLocalModules(".", config.LocalModules)
	if err != nil {
		return err
	}

	var replacedCount int
	var skippedCount int

	// Process each resolved module
	for _, resolved := range resolvedModules {
		if !resolved.IsDependency {
			fmt.Printf("Skipping module %s: no dependency found\n", resolved.ModuleInfo.Module.Path)
			skippedCount++
			continue
		}

		fmt.Printf("Adding replace directive for %s -> %s\n", resolved.ModuleInfo.Module.Path, resolved.LocalPath)

		// Execute go mod edit -replace
		err := commands.GoModEditReplace(resolved.ModuleInfo.Module.Path, resolved.LocalPath, nil)
		if err != nil {
			return fmt.Errorf("failed to add replace directive for %s: %w", resolved.ModuleInfo.Module.Path, err)
		}
		replacedCount++
	}

	if replacedCount == 0 {
		fmt.Printf("No modules were replaced\n")
	} else {
		fmt.Printf("Successfully added replace directives for %d module(s)\n", replacedCount)
	}
	if skippedCount > 0 {
		fmt.Printf("Skipped %d module(s)\n", skippedCount)
	}

	return nil
}
