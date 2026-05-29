package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	if err := handle(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle() error {
	rootDir, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	fmt.Println("==> Building frontend")
	if err := cmd.Debug().Dir(rootDir).Run("go", "run", "./script/build-react"); err != nil {
		return fmt.Errorf("build-react failed: %w", err)
	}

	fmt.Println("==> Installing kool")
	if err := cmd.Debug().Dir(rootDir).Run("go", "install", "."); err != nil {
		return fmt.Errorf("go install kool failed: %w", err)
	}

	fmt.Println("kool installed successfully")
	return nil
}

func findRepoRoot() (string, error) {
	if root, err := findRepoRootFromSourceFile(); err == nil {
		return root, nil
	}
	if root, err := findRepoRootFromWorkingDir(); err == nil {
		return root, nil
	}
	if root, err := findRepoRootFromGit(); err == nil {
		return root, nil
	}
	return "", fmt.Errorf("cannot find repository root")
}

func findRepoRootFromSourceFile() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot inspect source path")
	}
	return findRepoRootFrom(filepath.Dir(file))
}

func findRepoRootFromWorkingDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return findRepoRootFrom(wd)
}

func findRepoRootFromGit() (string, error) {
	out, err := cmd.Output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func findRepoRootFrom(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("repository root not found from %s", start)
}
