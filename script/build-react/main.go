package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

type frontend struct {
	name string
	dir  string
}

type options struct {
	frontend    string
	manager     string
	skipInstall bool
}

var frontends = []frontend{
	{
		name: "preview",
		dir:  filepath.Join("tools", "preview", "viewer", "react"),
	},
	{
		name: "web",
		dir:  filepath.Join("tools", "web", "react"),
	},
}

const help = `Usage:
  go run ./script/build-react [flags]

Flags:
  --frontend string     frontend to build: all, preview, or web (default "all")
  --manager string      package manager: auto, pnpm, bun, npm, or yarn (default "auto")
  --skip-install        skip dependency installation

Examples:
  go run ./script/build-react
  go run ./script/build-react --frontend preview
  go run ./script/build-react --skip-install
`

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}

	rootDir, err := findRepoRoot()
	if err != nil {
		return err
	}

	selected, err := selectFrontends(opts.frontend)
	if err != nil {
		return err
	}

	fmt.Printf("Using repository root: %s\n", rootDir)
	if opts.manager != "auto" {
		fmt.Printf("Using package manager: %s\n", opts.manager)
	}

	for _, item := range selected {
		if err := buildFrontend(rootDir, item, opts.manager, opts.skipInstall); err != nil {
			return err
		}
	}

	fmt.Println("React frontend files are ready.")
	return nil
}

func parseOptions(args []string) (options, error) {
	var opts options
	fs := flag.NewFlagSet("build-react", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&opts.frontend, "frontend", "all", "frontend to build: all, preview, or web")
	fs.StringVar(&opts.manager, "manager", "auto", "package manager: auto, pnpm, bun, npm, or yarn")
	fs.BoolVar(&opts.skipInstall, "skip-install", false, "skip dependency installation")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, help)
	}

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	if fs.NArg() != 0 {
		return options{}, fmt.Errorf("unrecognized extra args: %s", strings.Join(fs.Args(), " "))
	}
	return opts, nil
}

func selectFrontends(name string) ([]frontend, error) {
	switch name {
	case "", "all":
		return frontends, nil
	case "preview", "web":
		for _, item := range frontends {
			if item.name == name {
				return []frontend{item}, nil
			}
		}
	}
	return nil, fmt.Errorf("unknown frontend %q: expected all, preview, or web", name)
}

func buildFrontend(rootDir string, item frontend, managerPref string, skipInstall bool) error {
	frontendDir := filepath.Join(rootDir, item.dir)
	if !fileExists(filepath.Join(frontendDir, "package.json")) {
		return fmt.Errorf("%s frontend package.json not found at %s", item.name, frontendDir)
	}

	manager, err := npm.Resolve(frontendDir, managerPref)
	if err != nil {
		return err
	}

	fmt.Printf("\n==> Building %s frontend (%s)\n", item.name, item.dir)
	if managerPref == "auto" {
		fmt.Printf("Using package manager: %s\n", manager)
	}

	if !skipInstall {
		args := npm.InstallArgs(manager, npm.InstallOptions{})
		fmt.Printf("Installing dependencies: %s %s\n", manager, strings.Join(args, " "))
		if err := run(frontendDir, string(manager), args...); err != nil {
			return fmt.Errorf("%s dependency install failed: %w", item.name, err)
		}
	}

	fmt.Printf("Running build: %s run build\n", manager)
	if err := run(frontendDir, string(manager), "run", "build"); err != nil {
		return fmt.Errorf("%s build failed: %w", item.name, err)
	}

	distDir := filepath.Join(frontendDir, "dist")
	if err := verifyDist(item.name, distDir); err != nil {
		return err
	}
	if err := ensurePlaceholder(distDir); err != nil {
		return fmt.Errorf("%s placeholder failed: %w", item.name, err)
	}

	fmt.Printf("%s frontend ready: %s\n", item.name, distDir)
	return nil
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func verifyDist(name string, distDir string) error {
	indexPath := filepath.Join(distDir, "index.html")
	indexInfo, err := os.Stat(indexPath)
	if err != nil {
		return fmt.Errorf("%s build did not produce %s: %w", name, indexPath, err)
	}
	if indexInfo.IsDir() || indexInfo.Size() == 0 {
		return fmt.Errorf("%s build produced an empty or invalid index.html at %s", name, indexPath)
	}

	assetsDir := filepath.Join(distDir, "assets")
	entries, err := os.ReadDir(assetsDir)
	if err != nil {
		return fmt.Errorf("%s build did not produce %s: %w", name, assetsDir, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("%s build produced no frontend assets in %s", name, assetsDir)
	}

	hasFile := false
	for _, entry := range entries {
		if !entry.IsDir() {
			hasFile = true
			break
		}
	}
	if !hasFile {
		return fmt.Errorf("%s build produced no asset files in %s", name, assetsDir)
	}

	return nil
}

func ensurePlaceholder(distDir string) error {
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(distDir, "placeholder.txt"), nil, 0644)
}

func findRepoRoot() (string, error) {
	if root, err := findRepoRootFromWorkingDir(); err == nil {
		return root, nil
	}
	if root, err := findRepoRootFromSourceFile(); err == nil {
		return root, nil
	}
	if root, err := findRepoRootFromGit(); err == nil {
		return root, nil
	}
	return "", errors.New("cannot find repository root")
}

func findRepoRootFromWorkingDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return findRepoRootFrom(wd)
}

func findRepoRootFromSourceFile() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("cannot inspect source path")
	}
	return findRepoRootFrom(filepath.Dir(file))
}

func findRepoRootFromGit() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", errors.New("git returned empty repository root")
	}
	if err := validateRepoRoot(root); err != nil {
		return "", err
	}
	return root, nil
}

func findRepoRootFrom(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if err := validateRepoRoot(dir); err == nil {
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

func validateRepoRoot(dir string) error {
	goModPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	if !strings.Contains(string(data), "module github.com/xhd2015/kool") {
		return fmt.Errorf("%s is not the kool repository", goModPath)
	}
	for _, item := range frontends {
		if !fileExists(filepath.Join(dir, item.dir, "package.json")) {
			return fmt.Errorf("%s package.json is missing", item.name)
		}
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
