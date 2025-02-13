package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed frontend_template
var templateFS embed.FS

//go:embed server_template
var serverTemplateFS embed.FS

func create(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: kool create <TEMPLATE> <project-name>\nTEMPLATE: frontend, server")
	}

	template := args[0]

	if template == "" {
		return fmt.Errorf("requires template name, e.g. kool create frontend <project-name>")
	}
	if template != "frontend" && template != "server" {
		return fmt.Errorf("unsupported template: %s", template)
	}

	projectName := args[1]
	if projectName == "" {
		return fmt.Errorf("requires project name, e.g. kool create frontend <project-name>")
	}

	targetPath := filepath.Join(".", projectName)

	// Check if target directory already exists
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", projectName)
	}

	// Create the target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	srcFS := templateFS
	srcRoot := "frontend_template"
	if template == "server" {
		srcFS = serverTemplateFS
		srcRoot = "server_template"
	}
	// Copy template files to new project
	err := fs.WalkDir(srcFS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from template root
		relPath, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		targetFilePath := filepath.Join(targetPath, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetFilePath, 0755)
		}

		// Copy file contents
		content, err := srcFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %v", path, err)
		}

		if err := os.WriteFile(targetFilePath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %v", targetFilePath, err)
		}

		return nil
	})

	if err != nil {
		// Clean up on failure
		os.RemoveAll(targetPath)
		return fmt.Errorf("failed to copy template: %v", err)
	}

	// if server, rename go.mod.template to go.mod
	if template == "server" {
		err = os.Rename(filepath.Join(targetPath, "go.mod.template"), filepath.Join(targetPath, "go.mod"))
		if err != nil {
			return fmt.Errorf("failed to rename go.mod.template to go.mod: %v", err)
		}

		// if current project has remote github url, set to go mod
		suggstedGoModPath, _ := suggestGoModPath(targetPath)
		if suggstedGoModPath != "" {
			// replace defaultMod with suggstedGoModPath in go.mod, main.go
			files := []string{"go.mod", "main.go"}
			for _, file := range files {
				content, err := os.ReadFile(filepath.Join(targetPath, file))
				if err != nil {
					return fmt.Errorf("failed to read %s: %v", file, err)
				}
				content = bytes.ReplaceAll(content, []byte(defaultMod), []byte(suggstedGoModPath))
				err = os.WriteFile(filepath.Join(targetPath, file), content, 0644)
				if err != nil {
					return fmt.Errorf("failed to write %s: %v", file, err)
				}
			}
		}
	}

	fmt.Printf("Successfully created new project: %s\n", projectName)
	return nil
}

const defaultMod = "github.com/xhd2015/kool/server_template"

func suggestGoModPath(dir string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	data, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote url: %v", err)
	}

	output := strings.TrimSpace(string(data))

	// remove .git suffix
	remoteURL := strings.TrimSuffix(output, ".git")

	// example: ssh://git@github.com/xhd2015/kool
	if strings.HasPrefix(remoteURL, "ssh://git@github.com") {
		githubRepo := strings.TrimPrefix(remoteURL, "ssh://git@")

		// get relative path to git root
		subPaths, subPathsErr := getRelativePathToGitRoot(dir)
		if subPathsErr == nil {
			return githubRepo + "/" + strings.Join(subPaths, "/"), nil
		} else {
			return githubRepo, nil
		}
	}

	return "", nil
}

func getRelativePathToGitRoot(dir string) ([]string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-prefix")
	cmd.Dir = dir
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %v", err)
	}

	prefix := strings.TrimSpace(string(data))
	if prefix == "" {
		return []string{}, nil
	}

	// git always uses forward slashes, trim trailing slash if any
	prefix = strings.TrimSuffix(prefix, "/")
	return strings.Split(prefix, "/"), nil
}
