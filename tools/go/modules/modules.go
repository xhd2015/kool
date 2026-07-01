package modules

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	scanpkg "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/kool/tools/git/tag"
	"github.com/xhd2015/less-flags"
)

const help = `
kool go modules lists Go module directories under the current directory.

Usage: kool go modules [OPTIONS] [COMMAND]

Commands:
  ls-files           list files owned by a module
  update-local-deps  tag local modules and update local dependency versions

Options:
  --dir <dir>        root directory, default is current directory
  --list             stream "<dir> <module-path>" lines in walk order (no tags)
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
	var list bool
	args, err := parseLeadingModulesFlags(args, &dir, &noTags, &list)
	if err != nil {
		return err
	}
	if dir == "" {
		dir = "."
	}

	if len(args) > 0 {
		switch args[0] {
		case "ls-files":
			if list {
				return fmt.Errorf("--list is not supported with ls-files")
			}
			return handleLsFiles(w, dir, args[1:])
		case "update-local-deps":
			if noTags {
				return fmt.Errorf("--no-tags is not supported with update-local-deps")
			}
			if list {
				return fmt.Errorf("--list is not supported with update-local-deps")
			}
			return handleUpdateLocalDeps(w, dir, args[1:])
		case "help", "--help", "-h":
			fmt.Fprint(w, strings.TrimPrefix(help, "\n"))
			return nil
		}
	}

	if list && noTags {
		return fmt.Errorf("--list is not supported with --no-tags")
	}
	if list {
		return handleList(w, dir)
	}

	return handleDefault(w, dir, noTags, args)
}

func parseLeadingModulesFlags(args []string, dir *string, noTags *bool, list *bool) ([]string, error) {
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
		case arg == "--list":
			*list = true
			args = args[1:]
		default:
			return args, nil
		}
	}
	return args, nil
}

// handleList streams "<dir> <module-path>" lines to w in walk order (unsorted),
// one per Go module found under dir, delegating the walk to scan.ScanStream.
// Each line is flushed before the next module is emitted. dir is "." for the
// root module and the plain slash-relative path for sub-directories (no "./").
func handleList(w io.Writer, dir string) error {
	return scanpkg.ScanStream(dir, scanpkg.Options{}, func(m scanpkg.Module) error {
		_, err := fmt.Fprintln(w, m.Dir+" "+m.Path)
		return err
	})
}

func handleDefault(w io.Writer, dir string, noTags bool, args []string) error {
	var list bool
	args, err := lessflags.
		String("--dir", &dir).
		Bool("--list", &list).
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
	if list && noTags {
		return fmt.Errorf("--list is not supported with --no-tags")
	}
	if list {
		return handleList(w, dir)
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

	scanned, err := scanpkg.Scan(root, scanpkg.Options{})
	if err != nil {
		return nil, err
	}

	modules := make([]Module, 0, len(scanned))
	for _, sm := range scanned {
		module, err := convertScanModule(absRoot, sm, opts)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}

	// Scan already returns modules sorted by Dir; preserve that order.
	fillDependencies(modules)
	return modules, nil
}

// convertScanModule turns a scan.Module (core walk + go.mod read) into kool's
// richer Module, adding the latest-tag lookup and the requirePaths used by
// dependency filling. The walk + skip rules + go.mod parsing live in the scan
// package; kool only layers on its own annotations.
func convertScanModule(absRoot string, sm scanpkg.Module, opts FindOptions) (Module, error) {
	module := Module{
		Dir:      sm.Dir,
		Path:     sm.Path,
		Requires: convertRequires(sm.Requires),
		Replaces: convertReplaces(sm.Replaces),
	}
	requirePaths := make([]string, 0, len(sm.Requires))
	for _, req := range sm.Requires {
		requirePaths = append(requirePaths, req.Path)
	}
	module.requirePaths = requirePaths

	if !opts.NoTags {
		dir := absRoot
		if sm.Dir != "." && sm.Dir != "" {
			dir = filepath.Join(absRoot, filepath.FromSlash(sm.Dir))
		}
		module.LatestTag, module.LatestTagKnown = findLatestModuleTag(dir)
	}
	return module, nil
}

func convertRequires(reqs []scanpkg.ModuleRequire) []ModuleRequire {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]ModuleRequire, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, ModuleRequire{Path: r.Path, Version: r.Version})
	}
	return out
}

func convertReplaces(reps []scanpkg.ModuleReplace) []ModuleReplace {
	if len(reps) == 0 {
		return nil
	}
	out := make([]ModuleReplace, 0, len(reps))
	for _, r := range reps {
		out = append(out, ModuleReplace{
			OldPath:    r.OldPath,
			NewPath:    r.NewPath,
			NewVersion: r.NewVersion,
		})
	}
	return out
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
