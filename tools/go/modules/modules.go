package modules

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"golang.org/x/mod/modfile"
)

const help = `
kool go modules lists Go module directories under the current directory.

Usage: kool go modules [OPTIONS]

Options:
  --dir <dir>          root directory, default is current directory
  -h,--help            show help message
`

func Handle(args []string) error {
	var dir string
	args, err := flags.
		String("--dir", &dir).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	if dir == "" {
		dir = "."
	}

	modules, err := Find(dir)
	if err != nil {
		return err
	}
	return Render(os.Stdout, modules)
}

type Module struct {
	Dir     string
	Path    string
	Depends []string

	requirePaths []string
}

func Find(root string) ([]Module, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	ignoreChecker := newGitIgnoreChecker(absRoot)

	var modules []Module
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		if path != absRoot {
			switch d.Name() {
			case ".git", "vendor":
				return filepath.SkipDir
			}

			ignored, err := ignoreChecker.IsIgnored(path)
			if err != nil {
				return err
			}
			if ignored {
				return filepath.SkipDir
			}
		}

		hasGoMod, err := hasGoMod(path)
		if err != nil {
			return err
		}
		if hasGoMod {
			rel, err := filepath.Rel(absRoot, path)
			if err != nil {
				return err
			}
			module, err := readModule(path, filepath.ToSlash(rel))
			if err != nil {
				return err
			}
			modules = append(modules, module)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Dir < modules[j].Dir
	})
	fillDependencies(modules)
	return modules, nil
}

func hasGoMod(dir string) (bool, error) {
	info, err := os.Stat(filepath.Join(dir, "go.mod"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func readModule(dir string, rel string) (Module, error) {
	goModPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return Module{}, err
	}

	module := Module{
		Dir:  rel,
		Path: modfile.ModulePath(data),
	}

	modFile, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		modFile, err = modfile.ParseLax(goModPath, data, nil)
		if err != nil {
			return module, nil
		}
	}
	if modFile.Module != nil && modFile.Module.Mod.Path != "" {
		module.Path = modFile.Module.Mod.Path
	}

	requirePaths := make([]string, 0, len(modFile.Require))
	for _, req := range modFile.Require {
		requirePaths = append(requirePaths, req.Mod.Path)
	}
	module.requirePaths = requirePaths

	return module, nil
}

func fillDependencies(modules []Module) {
	modulePathDirs := make(map[string]string, len(modules))
	for _, module := range modules {
		if _, ok := modulePathDirs[module.Path]; !ok {
			modulePathDirs[module.Path] = module.Dir
		}
	}

	for i := range modules {
		depSet := make(map[string]struct{})
		for _, requirePath := range modules[i].requirePaths {
			depDir, ok := modulePathDirs[requirePath]
			if !ok || depDir == modules[i].Dir {
				continue
			}
			depSet[depDir] = struct{}{}
		}
		modules[i].Depends = modules[i].Depends[:0]
		for depDir := range depSet {
			modules[i].Depends = append(modules[i].Depends, depDir)
		}
		sort.Strings(modules[i].Depends)
	}
}

type gitIgnoreChecker struct {
	root        string
	ignoredDirs map[string]struct{}
	fallback    bool
}

func newGitIgnoreChecker(root string) gitIgnoreChecker {
	if _, err := exec.LookPath("git"); err != nil {
		return gitIgnoreChecker{}
	}

	cmd := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	if err != nil {
		return gitIgnoreChecker{}
	}
	if strings.TrimSpace(string(output)) != "true" {
		return gitIgnoreChecker{}
	}

	cmd = exec.Command("git", "-C", root, "ls-files", "--others", "--ignored", "--exclude-standard", "--directory", "-z", "--", ".")
	output, err = cmd.Output()
	if err != nil {
		return gitIgnoreChecker{}
	}

	ignoredDirs := make(map[string]struct{})
	for _, entry := range strings.Split(string(output), "\x00") {
		entry = filepath.ToSlash(entry)
		if !strings.HasSuffix(entry, "/") {
			continue
		}
		entry = strings.TrimSuffix(entry, "/")
		entry = strings.TrimPrefix(entry, "./")
		entry = filepath.ToSlash(filepath.Clean(entry))
		if entry == "." || entry == "" {
			continue
		}
		ignoredDirs[entry] = struct{}{}
	}

	return gitIgnoreChecker{
		root:        root,
		ignoredDirs: ignoredDirs,
		fallback:    true,
	}
}

func (c gitIgnoreChecker) IsIgnored(path string) (bool, error) {
	if !c.fallback {
		return false, nil
	}

	rel, err := filepath.Rel(c.root, path)
	if err != nil {
		return false, err
	}
	if rel == "." {
		return false, nil
	}
	rel = filepath.ToSlash(filepath.Clean(rel))

	for current := rel; current != "." && current != ""; {
		if _, ok := c.ignoredDirs[current]; ok {
			return true, nil
		}
		next := filepath.ToSlash(filepath.Dir(current))
		if next == current {
			break
		}
		current = next
	}
	return c.checkIgnore(rel)
}

func (c gitIgnoreChecker) checkIgnore(rel string) (bool, error) {
	cmd := exec.Command("git", "-C", c.root, "check-ignore", "-q", "--", rel+"/")
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

type treeNode struct {
	name     string
	module   *Module
	children map[string]*treeNode
}

func newTreeNode(name string) *treeNode {
	return &treeNode{
		name:     name,
		children: make(map[string]*treeNode),
	}
}

func Render(w io.Writer, modules []Module) error {
	root := newTreeNode(".")
	for i := range modules {
		addModule(root, &modules[i])
	}

	if _, err := fmt.Fprintln(w, root.name); err != nil {
		return err
	}
	return renderChildren(w, root, "")
}

func addModule(root *treeNode, module *Module) {
	dir := module.Dir
	dir = filepath.ToSlash(filepath.Clean(dir))
	if dir == "" || dir == "." {
		root.module = module
		return
	}

	node := root
	for _, part := range strings.Split(dir, "/") {
		if part == "" || part == "." {
			continue
		}
		child := node.children[part]
		if child == nil {
			child = newTreeNode(part)
			node.children[part] = child
		}
		node = child
	}
	node.module = module
}

func renderChildren(w io.Writer, node *treeNode, prefix string) error {
	entries := make([]treeEntry, 0, len(node.children)+1)
	for _, child := range node.children {
		entries = append(entries, treeEntry{name: child.name, node: child})
	}
	if node.module != nil {
		entries = append(entries, treeEntry{name: "go.mod", module: node.module})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	for i, entry := range entries {
		last := i == len(entries)-1
		connector := "├── "
		nextPrefix := prefix + "│   "
		if last {
			connector = "└── "
			nextPrefix = prefix + "    "
		}

		if _, err := fmt.Fprintln(w, prefix+connector+entry.name); err != nil {
			return err
		}
		if err := renderDependencyLines(w, entry.module, nextPrefix); err != nil {
			return err
		}
		if entry.node != nil {
			if err := renderChildren(w, entry.node, nextPrefix); err != nil {
				return err
			}
		}
	}
	return nil
}

type treeEntry struct {
	name   string
	node   *treeNode
	module *Module
}

func renderDependencyLines(w io.Writer, module *Module, prefix string) error {
	if module == nil {
		return nil
	}
	for i, dep := range module.Depends {
		connector := "├── "
		if i == len(module.Depends)-1 {
			connector = "└── "
		}
		if _, err := fmt.Fprintf(w, "%s%s(depends on) %s\n", prefix, connector, depGoModPath(dep)); err != nil {
			return err
		}
	}
	return nil
}

func depGoModPath(dir string) string {
	dir = filepath.ToSlash(filepath.Clean(dir))
	if dir == "." || dir == "" {
		return "go.mod"
	}
	return dir + "/go.mod"
}
