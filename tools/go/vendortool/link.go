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

const help = `
Usage: kool go vendor <command>

Commands:
  link <target_dir>
  link --all                           link all local modules from config
  unlink <target_dir>
  unlink --all                         unlink all local modules from config
`

// rename old path to {old path}__old, and create new symlink
func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kool go vendor <subcommand>, check kool go vendor --help")
	}
	cmd := args[0]
	args = args[1:]

	if cmd == "--help" || cmd == "help" {
		fmt.Println(strings.TrimPrefix(help, "\n"))
		return nil
	}

	switch cmd {
	case "link":
		return HandleLink(args)
	case "unlink":
		return HandleUnlink(args)
	default:
		return fmt.Errorf("unrecognized command: %s", cmd)
	}
}

const linkHelp = `
Create symlink for a local module in the vendor directory,
which makes local development easier.

Usage: kool go vendor link <target_dir>
       kool go vendor link --all

Options:
  -h,--help            show help message
  --dir <dir>          set the directory to vendor
  -v,--verbose         show verbose info
  --all                link all local modules from config
  --show               show config file path

Examples:
  kool go vendor link ~/external/module
  kool go vendor link --all
`

func HandleLink(args []string) error {
	var dir string
	var verbose bool
	var all bool
	var show bool
	args, err := flags.
		String("--dir", &dir).
		Help("-h,--help", linkHelp).
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
		return LinkAll(dir, verbose)
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: kool go vendor link <target_dir>")
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

	// get the mod path
	modInfo, err := resolve.GetModuleInfo(targetDir)
	if err != nil {
		return err
	}
	if modInfo.Module.Path == "" {
		return fmt.Errorf("not a go module: %s", targetDir)
	}

	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}

	// Use the extracted function
	err = linkVendorModule(modInfo.Module.Path, absTargetDir, vendorDir, verbose)
	if err != nil {
		return err
	}

	fmt.Printf("unlink with:\n  kool go vendor unlink %s\n", targetDir)

	return nil
}

// linkVendorModule creates a vendor symlink for a single module
func linkVendorModule(modulePath, sourcePath, vendorDir string, verbose bool) error {
	localVendorDir := filepath.Join(vendorDir, modulePath)

	// Check if already exists as symlink
	isSym, err := fs.IsSymbolicLink(localVendorDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if %s is symlink: %w", localVendorDir, err)
	}
	if isSym {
		fmt.Printf("symlink %s already exists\n", modulePath)
		return nil
	}

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(localVendorDir), 0755); err != nil {
		return fmt.Errorf("failed to create vendor parent directories: %w", err)
	}

	// If directory exists (not symlink), rename it to __old
	if _, err := os.Stat(localVendorDir); err == nil {
		oldPath := localVendorDir + "__old"
		if _, err := os.Stat(oldPath); err == nil {
			return fmt.Errorf("backup path already exists: %s", oldPath)
		}
		if err := os.Rename(localVendorDir, oldPath); err != nil {
			return fmt.Errorf("failed to backup existing vendor dir: %w", err)
		}
		if verbose {
			fmt.Printf("Backed up existing vendor dir: %s -> %s\n", filepath.Base(localVendorDir), filepath.Base(oldPath))
		}
	}

	fmt.Printf("symlink %s\n", sourcePath)

	// Create symlink
	if err := os.Symlink(sourcePath, localVendorDir); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// LinkAll reads the config, gets the LocalModules list, and creates symlinks for dependencies
func LinkAll(dir string, verbose bool) error {
	config, err := goconfig.GetLocalModulesConfig()
	if err != nil {
		return err
	}

	if len(config.LocalModules) == 0 {
		fmt.Printf("No local modules configured\n")
		return nil
	}

	// Ensure vendor directory exists
	vendorDir := filepath.Join(dir, "vendor")
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		return fmt.Errorf("failed to create vendor directory: %w", err)
	}

	// Resolve all local modules and their dependency status
	_, resolvedModules, err := resolve.ResolveLocalModules(".", config.LocalModules)
	if err != nil {
		return err
	}

	var linkedCount int
	var skippedCount int

	// Process each resolved module
	for _, resolved := range resolvedModules {
		if !resolved.IsDependency {
			if verbose {
				fmt.Printf("Skipping module %s: no dependency found\n", resolved.ModuleInfo.Module.Path)
			}
			skippedCount++
			continue
		}

		err := linkVendorModule(resolved.ModuleInfo.Module.Path, resolved.LocalPath, vendorDir, verbose)
		if err != nil {
			return fmt.Errorf("failed to create vendor symlink for %s: %w", resolved.ModuleInfo.Module.Path, err)
		}
		linkedCount++
	}

	if linkedCount == 0 {
		fmt.Printf("No modules were linked\n")
	} else {
		fmt.Printf("Successfully created vendor symlinks for %d module(s)\n", linkedCount)
	}
	if skippedCount > 0 {
		fmt.Printf("Skipped %d module(s)\n", skippedCount)
	}

	return nil
}
