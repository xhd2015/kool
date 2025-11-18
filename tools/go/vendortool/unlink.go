package vendortool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/kool/tools/fs"
	goconfig "github.com/xhd2015/kool/tools/go/config"
	"github.com/xhd2015/kool/tools/go/resolve"
	"github.com/xhd2015/less-gen/flags"
)

const unlinkHelp = `
Remove the symlink for a local module in the vendor directory.

Usage: kool go vendor unlink <target_dir>
       kool go vendor unlink --all

Options:
  -h,--help            show help message
  --dir <dir>          set the directory to vendor
  -v,--verbose         show verbose info
  --all                remove all symlinks from config
  --show               show config file path

Examples:
  kool go vendor unlink ~/external/module
  kool go vendor unlink --all
`

func HandleUnlink(args []string) error {
	var dir string
	var verbose bool
	var all bool
	var show bool
	args, err := flags.
		String("--dir", &dir).
		Help("-h,--help", unlinkHelp).
		Bool("-v,--verbose", &verbose).
		Bool("--all", &all).
		Bool("--show", &show).
		Parse(args)
	if err != nil {
		return err
	}

	if all {
		if len(args) > 0 {
			return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
		}
		if show {
			return goconfig.ShowLocalModulesConfig()
		}
		return UnlinkAll(dir, verbose)
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: kool go vendor unlink <target_dir>")
	}
	targetDir := args[0]
	args = args[1:]
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	err = fs.ValidateIsDir(targetDir)
	if err != nil {
		return err
	}
	vendorDir := filepath.Join(dir, "vendor")
	err = fs.ValidateIsDir(vendorDir)
	if err != nil {
		return err
	}

	modInfo, err := resolve.GetModuleInfo(targetDir)
	if err != nil {
		return err
	}
	if modInfo.Module.Path == "" {
		return fmt.Errorf("not a go module: %s", targetDir)
	}

	// Use the extracted function
	err = unlinkVendorModule(modInfo.Module.Path, vendorDir, verbose)
	if err != nil {
		return err
	}

	return nil
}

// unlinkVendorModule removes a vendor symlink for a single module
func unlinkVendorModule(modulePath, vendorDir string, verbose bool) error {
	localVendorDir := filepath.Join(vendorDir, modulePath)

	// Check if it exists as symlink
	isSym, err := fs.IsSymbolicLink(localVendorDir)
	if err != nil {
		if os.IsNotExist(err) {
			if verbose {
				fmt.Printf("Vendor symlink does not exist for %s\n", modulePath)
			}
			return nil
		}
		return fmt.Errorf("failed to check if %s is symlink: %w", localVendorDir, err)
	}
	if !isSym {
		if verbose {
			fmt.Printf("Vendor directory is not a symlink for %s\n", modulePath)
		}
		return nil
	}

	fmt.Printf("Removing vendor symlink for %s\n", modulePath)

	// Remove symlink
	if err := os.Remove(localVendorDir); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	// Check if there's a backup to restore
	oldPath := localVendorDir + "__old"
	if _, err := os.Stat(oldPath); err == nil {
		if err := os.Rename(oldPath, localVendorDir); err != nil {
			fmt.Printf("Warning: failed to restore backup %s: %v\n", oldPath, err)
		} else if verbose {
			fmt.Printf("Restored backup: %s -> %s\n", filepath.Base(oldPath), filepath.Base(localVendorDir))
		}
	}

	return nil
}

// UnlinkAll reads the config, gets the LocalModules list, and removes symlinks for all configured modules
func UnlinkAll(dir string, verbose bool) error {
	config, err := goconfig.GetLocalModulesConfig()
	if err != nil {
		return err
	}

	if len(config.LocalModules) == 0 {
		fmt.Printf("No local modules configured\n")
		return nil
	}

	vendorDir := filepath.Join(dir, "vendor")
	if _, err := os.Stat(vendorDir); os.IsNotExist(err) {
		fmt.Printf("Vendor directory does not exist: %s\n", vendorDir)
		return nil
	}

	// Resolve all local modules to get their module paths
	_, resolvedModules, err := resolve.ResolveLocalModules(".", config.LocalModules)
	if err != nil {
		return err
	}

	var unlinkedCount int
	var skippedCount int

	// Process each resolved module
	for _, resolved := range resolvedModules {
		err := unlinkVendorModule(resolved.ModuleInfo.Module.Path, vendorDir, verbose)
		if err != nil {
			fmt.Printf("Failed to unlink module %s: %v\n", resolved.ModuleInfo.Module.Path, err)
			skippedCount++
			continue
		}
		unlinkedCount++
	}

	if unlinkedCount == 0 {
		fmt.Printf("No modules were unlinked\n")
	} else {
		fmt.Printf("Successfully removed vendor symlinks for %d module(s)\n", unlinkedCount)
	}
	if skippedCount > 0 {
		fmt.Printf("Failed to unlink %d module(s)\n", skippedCount)
	}

	return nil
}
