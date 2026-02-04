package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage:
  init_project [flags]

Flags:
  --module string   Module path for the new project (required)
  -v, --verbose     Verbose output
  -h, --help        Show help

Example:
  go run ./script/init_project --module github.com/myorg/myproject
`

const templateModulePath = "github.com/xhd2015/kool/tools/create/server_go_db_template"

func main() {
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	var verbose bool
	var modulePath string
	args, err := flags.
		Help("-h,--help", help).
		Bool("-v,--verbose", &verbose).
		String("--module", &modulePath).
		Parse(args)
	if err != nil {
		return err
	}

	if modulePath == "" {
		return fmt.Errorf("--module flag is required")
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	if verbose {
		fmt.Printf("Project root: %s\n", projectRoot)
		fmt.Printf("Replacing module path: %s -> %s\n", templateModulePath, modulePath)
	}

	err = replaceModulePath(projectRoot, templateModulePath, modulePath, verbose)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully initialized project with module: %s\n", modulePath)
	return nil
}

// replaceModulePath replaces all occurrences of oldModule with newModule
// in .go files and go.mod within the project directory
func replaceModulePath(projectRoot, oldModule, newModule string, verbose bool) error {
	return filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") && filepath.Base(path) != "go.mod" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		newContent := bytes.ReplaceAll(content, []byte(oldModule), []byte(newModule))
		if bytes.Equal(content, newContent) {
			return nil
		}

		if verbose {
			fmt.Printf("Updating: %s\n", path)
		}

		err = os.WriteFile(path, newContent, info.Mode())
		if err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}

		return nil
	})
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
