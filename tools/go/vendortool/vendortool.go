package vendortool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/kool/tools/fs"
	"github.com/xhd2015/kool/tools/go/resolve"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: kool go vendor <command>

Commands:
  link <target_dir>
  unlink <target_dir>
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

Options:
  -h,--help            show help message
  --dir <dir>          set the directory to vendor
  -v,--verbose         show verbose info

Examples:
  kool go vendor link ~/external/module
`

func HandleLink(args []string) error {
	var dir string
	var verbose bool
	args, err := flags.
		String("--dir", &dir).
		Help("-h,--help", linkHelp).
		Bool("-v,--verbose", &verbose).
		Parse(args)
	if err != nil {
		return err
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

	// check in the vendor dir
	localVendorDir := filepath.Join(vendorDir, modInfo.Module.Path)

	isSym, err := fs.IsSymbolicLink(localVendorDir)
	if err != nil {
		return err
	}
	if isSym {
		fmt.Fprintf(os.Stderr, "local vendor dir is already a symlink: %s\n", localVendorDir)
		return nil
	}
	// check if the vendor dir contains the mod path

	err = fs.ValidateIsDir(localVendorDir)
	if err != nil {
		return err
	}

	oldPath := localVendorDir + "__old"

	err = fs.ValidateNotExists(oldPath)
	if err != nil {
		return err
	}

	err = os.Rename(localVendorDir, oldPath)
	if err != nil {
		return fmt.Errorf("rename %s -> %s:  %w", localVendorDir, oldPath, err)
	}
	fmt.Fprintf(os.Stderr, "renamed %s -> %s\n", filepath.Base(localVendorDir), filepath.Base(oldPath))

	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	err = os.Symlink(absTargetDir, localVendorDir)
	if err != nil {
		return err
	}

	fmt.Printf("created symlink vendor/.../%s -> %s, unlink with:\n  kool go vendor unlink %s\n", filepath.Base(localVendorDir), targetDir, targetDir)

	return nil
}

const unlinkHelp = `
Remove the symlink for a local module in the vendor directory.

Usage: kool go vendor unlink <target_dir>

Options:
  -h,--help            show help message
  --dir <dir>          set the directory to vendor
  -v,--verbose         show verbose info
  --all                remove all symlinks in the vendor directory
`

func HandleUnlink(args []string) error {
	var dir string
	var verbose bool
	args, err := flags.
		String("--dir", &dir).
		Help("-h,--help", unlinkHelp).
		Bool("-v,--verbose", &verbose).
		Parse(args)
	if err != nil {
		return err
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

	localVendorDir := filepath.Join(vendorDir, modInfo.Module.Path)

	isSym, err := fs.IsSymbolicLink(localVendorDir)
	if err != nil {
		return err
	}
	if !isSym {
		fmt.Fprintf(os.Stderr, "local vendor dir is not a symlink: %s\n", localVendorDir)
		return nil
	}

	err = os.Remove(localVendorDir)
	if err != nil {
		return err
	}

	// if there is a {name}__old, rename it to {name}
	oldPath := localVendorDir + "__old"
	err = fs.ValidateIsDir(oldPath)
	if err != nil {
		return err
	}
	err = os.Rename(oldPath, localVendorDir)
	if err != nil {
		return err
	}

	return nil
}
