package replace

import (
	"fmt"

	goconfig "github.com/xhd2015/kool/tools/go/config"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	gotoolreplace "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
)

// Replace checks if the given directory exists and contains a valid Go module,
// returns the absolute path of the directory and the module path.
func Replace(dir string) (absDir string, modulePath string, err error) {
	return gotoolreplace.Replace(dir)
}

// ReplaceAll reads the config, gets the LocalModules list, and adds replace directives for dependencies.
func ReplaceAll() error {
	config, err := goconfig.GetLocalModulesConfig()
	if err != nil {
		return err
	}

	if len(config.LocalModules) == 0 {
		fmt.Printf("No local modules configured\n")
		return nil
	}

	_, resolvedModules, err := resolve.ResolveLocalModules(".", config.LocalModules)
	if err != nil {
		return err
	}

	var replacedCount int
	var skippedCount int

	for _, resolved := range resolvedModules {
		if !resolved.IsDependency {
			fmt.Printf("Skipping module %s: no dependency found\n", resolved.ModuleInfo.Module.Path)
			skippedCount++
			continue
		}

		fmt.Printf("Adding replace directive for %s -> %s\n", resolved.ModuleInfo.Module.Path, resolved.LocalPath)

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