package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func ListModuleFiles(root string, moduleDir string) ([]string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	absRoot, err = filepath.EvalSymlinks(absRoot)
	if err != nil {
		return nil, err
	}
	if _, err := requireGitRepo(absRoot); err != nil {
		return nil, err
	}

	moduleDir = cleanModuleDir(moduleDir)
	modules, err := FindWithOptions(absRoot, FindOptions{NoTags: true})
	if err != nil {
		return nil, err
	}
	module, ok := findModule(modules, moduleDir)
	if !ok {
		return nil, fmt.Errorf("module %s was not found under %s; available modules: %s", moduleDir, root, moduleDirList(modules))
	}

	files, err := gitWorkspaceFiles(absRoot, module.Dir)
	if err != nil {
		return nil, err
	}
	nestedModuleDirs := nestedModuleDirsFor(module.Dir, modules)

	var result []string
	seen := make(map[string]bool, len(files))
	for _, file := range files {
		file = filepath.ToSlash(filepath.Clean(file))
		if seen[file] {
			continue
		}
		seen[file] = true
		if !isFileInModule(file, module.Dir) {
			continue
		}
		if fileDoesNotExistOrIsDir(absRoot, file) {
			continue
		}
		if isUnderNestedModule(file, nestedModuleDirs) {
			continue
		}
		result = append(result, file)
	}
	sort.Strings(result)
	return result, nil
}

func cleanModuleDir(moduleDir string) string {
	moduleDir = strings.TrimSpace(moduleDir)
	if moduleDir == "" || moduleDir == "." {
		return "."
	}
	return filepath.ToSlash(filepath.Clean(moduleDir))
}

func findModule(modules []Module, moduleDir string) (Module, bool) {
	for _, module := range modules {
		if module.Dir == moduleDir {
			return module, true
		}
	}
	return Module{}, false
}

func moduleDirList(modules []Module) string {
	dirs := make([]string, 0, len(modules))
	for _, module := range modules {
		dirs = append(dirs, module.Dir)
	}
	sort.Strings(dirs)
	return strings.Join(dirs, ", ")
}

func gitWorkspaceFiles(root string, moduleDir string) ([]string, error) {
	pathspec := "."
	if moduleDir != "." && moduleDir != "" {
		pathspec = moduleDir
	}
	args := []string{"ls-files", "--cached", "--others", "--exclude-standard", "-z", "--", pathspec}
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, formatCmdError("git", args, err, output)
	}

	var files []string
	for _, file := range strings.Split(string(output), "\x00") {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		files = append(files, filepath.ToSlash(filepath.Clean(file)))
	}
	return files, nil
}

func isFileInModule(file string, moduleDir string) bool {
	if moduleDir == "." || moduleDir == "" {
		return true
	}
	return file == moduleDir || strings.HasPrefix(file, moduleDir+"/")
}

func nestedModuleDirsFor(moduleDir string, modules []Module) []string {
	var dirs []string
	for _, module := range modules {
		if module.Dir == moduleDir {
			continue
		}
		if isNestedModuleDir(module.Dir, moduleDir) {
			dirs = append(dirs, module.Dir)
		}
	}
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) < len(dirs[j])
	})
	return dirs
}

func isNestedModuleDir(candidate string, parent string) bool {
	if parent == "." || parent == "" {
		return candidate != "." && candidate != ""
	}
	return strings.HasPrefix(candidate, parent+"/")
}

func fileDoesNotExistOrIsDir(root string, file string) bool {
	info, err := os.Lstat(filepath.Join(root, filepath.FromSlash(file)))
	if err != nil {
		return true
	}
	return info.IsDir()
}
