package modules

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/kool/tools/git/tag"
	"github.com/xhd2015/less-flags"
	"golang.org/x/mod/modfile"
)

const help = `
kool go modules lists Go module directories under the current directory.

Usage: kool go modules [OPTIONS] [COMMAND]

Commands:
  ls-files           list files owned by a module
  update-local-deps  tag local modules and update local dependency versions

Options:
  --dir <dir>        root directory, default is current directory
  --no-tags          hide latest tag annotations
  -h,--help          show help message
`

const lsFilesHelp = `
Usage: kool go modules ls-files [OPTIONS]

List files owned by a Go module, including untracked files and excluding
git-ignored files and nested module directories.

Options:
  --dir <dir>        root directory, default is current directory
  --module <module>  module directory, such as "." or "types"
  -h,--help          show help message
`

const updateLocalDepsHelp = `
Usage: kool go modules update-local-deps [OPTIONS]

Tag local modules and update local dependency versions.

Options:
  --dir <dir>        root directory, default is current directory
  --dry-run          print expected output without touching anything
  -h,--help          show help message
`

func Handle(args []string) error {
	return handle(os.Stdout, args)
}

func handle(w io.Writer, args []string) error {
	var dir string
	var noTags bool
	args, err := parseLeadingModulesFlags(args, &dir, &noTags)
	if err != nil {
		return err
	}
	if dir == "" {
		dir = "."
	}

	if len(args) > 0 {
		switch args[0] {
		case "ls-files":
			return handleLsFiles(w, dir, args[1:])
		case "update-local-deps":
			if noTags {
				return fmt.Errorf("--no-tags is not supported with update-local-deps")
			}
			return handleUpdateLocalDeps(w, dir, args[1:])
		case "help", "--help", "-h":
			fmt.Fprint(w, strings.TrimPrefix(help, "\n"))
			return nil
		}
	}

	return handleDefault(w, dir, noTags, args)
}

func parseLeadingModulesFlags(args []string, dir *string, noTags *bool) ([]string, error) {
	for len(args) > 0 {
		arg := args[0]
		switch {
		case arg == "-h" || arg == "--help":
			return []string{arg}, nil
		case arg == "--dir":
			if len(args) < 2 {
				return nil, fmt.Errorf("--dir requires a value")
			}
			*dir = args[1]
			args = args[2:]
		case strings.HasPrefix(arg, "--dir="):
			*dir = strings.TrimPrefix(arg, "--dir=")
			args = args[1:]
		case arg == "--no-tags":
			*noTags = true
			args = args[1:]
		default:
			return args, nil
		}
	}
	return args, nil
}

func handleDefault(w io.Writer, dir string, noTags bool, args []string) error {
	args, err := lessflags.
		String("--dir", &dir).
		Bool("--no-tags", &noTags).
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

	modules, err := FindWithOptions(dir, FindOptions{NoTags: noTags})
	if err != nil {
		return err
	}
	return Render(w, modules)
}

func handleLsFiles(w io.Writer, dir string, args []string) error {
	var moduleDir string
	args, err := lessflags.
		String("--dir", &dir).
		String("--module", &moduleDir).
		Help("-h,--help", lsFilesHelp).
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
	if moduleDir == "" {
		return fmt.Errorf("--module is required")
	}

	files, err := ListModuleFiles(dir, moduleDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if _, err := fmt.Fprintln(w, file); err != nil {
			return err
		}
	}
	return nil
}

func handleUpdateLocalDeps(w io.Writer, dir string, args []string) error {
	var dryRun bool
	args, err := lessflags.
		String("--dir", &dir).
		Bool("--dry-run", &dryRun).
		Help("-h,--help", updateLocalDepsHelp).
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
	return UpdateLocalDepsAndRender(w, dir, dryRun)
}

type Module struct {
	Dir            string
	Path           string
	Depends        []string
	LatestTag      string
	LatestTagKnown bool

	Requires []ModuleRequire
	Replaces []ModuleReplace

	requirePaths []string
}

type ModuleRequire struct {
	Path    string
	Version string
}

type ModuleReplace struct {
	OldPath    string
	NewPath    string
	NewVersion string
}

type FindOptions struct {
	NoTags bool
}

func Find(root string) ([]Module, error) {
	return FindWithOptions(root, FindOptions{})
}

func FindWithOptions(root string, opts FindOptions) ([]Module, error) {
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
			module, err := readModule(path, filepath.ToSlash(rel), opts)
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

func readModule(dir string, rel string, opts FindOptions) (Module, error) {
	goModPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return Module{}, err
	}

	module := Module{
		Dir:  rel,
		Path: modfile.ModulePath(data),
	}
	if !opts.NoTags {
		module.LatestTag, module.LatestTagKnown = findLatestModuleTag(dir)
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
		module.Requires = append(module.Requires, ModuleRequire{
			Path:    req.Mod.Path,
			Version: req.Mod.Version,
		})
	}
	module.requirePaths = requirePaths
	for _, repl := range modFile.Replace {
		module.Replaces = append(module.Replaces, ModuleReplace{
			OldPath:    repl.Old.Path,
			NewPath:    repl.New.Path,
			NewVersion: repl.New.Version,
		})
	}

	return module, nil
}

func findLatestModuleTag(dir string) (string, bool) {
	versionPrefix, err := tag.GetVersionPrefix(dir)
	if err != nil {
		return "", false
	}
	latestTag, err := tag.GetLatestVersionTag(dir, versionPrefix)
	if err != nil {
		if errors.Is(err, tag.ErrNoTag) {
			return "", true
		}
		return "", false
	}
	return latestTag, true
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
	return RenderAnnotated(w, modules, nil)
}

type ModuleAnnotation struct {
	UpdatedDeps []DependencyAnnotation
	PreviousTag string
	NewTag      string
}

type DependencyAnnotation struct {
	Dir            string
	ModulePath     string
	OldVersion     string
	NewVersion     string
	RemovedReplace bool
}

func RenderAnnotated(w io.Writer, modules []Module, annotations map[string]ModuleAnnotation) error {
	root := newTreeNode(".")
	moduleByDir := make(map[string]*Module, len(modules))
	for i := range modules {
		addModule(root, &modules[i])
		moduleByDir[modules[i].Dir] = &modules[i]
	}

	if _, err := fmt.Fprintln(w, root.name); err != nil {
		return err
	}
	return renderChildren(w, root, "", annotations, moduleByDir)
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

func renderChildren(w io.Writer, node *treeNode, prefix string, annotations map[string]ModuleAnnotation, moduleByDir map[string]*Module) error {
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

		line := prefix + connector + entry.name
		if entry.module != nil {
			if annotation := formatModuleAnnotation(entry.module, annotations[entry.module.Dir]); annotation != "" {
				line += " " + annotation
			}
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
		if err := renderDependencyLines(w, entry.module, nextPrefix, moduleByDir); err != nil {
			return err
		}
		if entry.node != nil {
			if err := renderChildren(w, entry.node, nextPrefix, annotations, moduleByDir); err != nil {
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

func renderDependencyLines(w io.Writer, module *Module, prefix string, moduleByDir map[string]*Module) error {
	if module == nil {
		return nil
	}
	for i, dep := range module.Depends {
		connector := "├── "
		if i == len(module.Depends)-1 {
			connector = "└── "
		}
		line := fmt.Sprintf("%s%s(depends on) %s", prefix, connector, depGoModPath(dep))
		if version := dependencyVersion(module, dep, moduleByDir); version != "" {
			line += " [version: " + version + "]"
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func dependencyVersion(module *Module, depDir string, moduleByDir map[string]*Module) string {
	depModule := moduleByDir[depDir]
	if depModule == nil {
		return ""
	}
	for _, req := range module.Requires {
		if req.Path == depModule.Path {
			return req.Version
		}
	}
	return ""
}

func depGoModPath(dir string) string {
	dir = filepath.ToSlash(filepath.Clean(dir))
	if dir == "." || dir == "" {
		return "go.mod"
	}
	return dir + "/go.mod"
}

func formatModuleAnnotation(module *Module, annotation ModuleAnnotation) string {
	var parts []string
	if module != nil && module.LatestTagKnown {
		latestTag := module.LatestTag
		if latestTag == "" {
			latestTag = "<none>"
		}
		parts = append(parts, "latest tag: "+latestTag)
	}
	parts = appendModuleAnnotationParts(parts, annotation)
	if len(parts) == 0 {
		return ""
	}
	return "[" + strings.Join(parts, "; ") + "]"
}

func formatAnnotation(annotation ModuleAnnotation) string {
	parts := appendModuleAnnotationParts(nil, annotation)
	if len(parts) == 0 {
		return ""
	}
	return "[" + strings.Join(parts, "; ") + "]"
}

func appendModuleAnnotationParts(parts []string, annotation ModuleAnnotation) []string {
	for _, dep := range annotation.UpdatedDeps {
		label := depGoModPath(dep.Dir)
		if dep.OldVersion != "" && dep.OldVersion != dep.NewVersion {
			label += " " + dep.OldVersion + " -> " + dep.NewVersion
		} else {
			label += " -> " + dep.NewVersion
		}
		if dep.RemovedReplace {
			label += ", replace removed"
		}
		parts = append(parts, "updated: "+label)
	}
	if annotation.NewTag != "" {
		if annotation.PreviousTag != "" {
			parts = append(parts, "new tag: "+annotation.PreviousTag+" -> "+annotation.NewTag)
		} else {
			parts = append(parts, "new tag: "+annotation.NewTag)
		}
	}
	return parts
}
