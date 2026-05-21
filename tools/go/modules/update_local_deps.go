package modules

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	git_tag_next "github.com/xhd2015/kool/tools/git/git_tag_next"
	"github.com/xhd2015/kool/tools/git/tag"
)

func UpdateLocalDepsAndRender(w io.Writer, dir string, dryRun bool) error {
	absRoot, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	absRoot, err = filepath.EvalSymlinks(absRoot)
	if err != nil {
		return err
	}

	gitRoot, err := requireGitRepo(absRoot)
	if err != nil {
		return err
	}
	if err := requireCleanGitRepo(gitRoot); err != nil {
		return err
	}

	annotations, err := updateLocalDeps(absRoot, gitRoot, updateLocalDepsOptions{
		DryRun: dryRun,
	})
	if err != nil {
		return err
	}

	modules, err := Find(absRoot)
	if err != nil {
		return err
	}
	return RenderAnnotated(w, modules, annotations)
}

type updateLocalDepsOptions struct {
	DryRun bool
}

type updateLocalDepsState struct {
	plannedTags map[string]string
}

func newUpdateLocalDepsState() *updateLocalDepsState {
	return &updateLocalDepsState{
		plannedTags: make(map[string]string),
	}
}

func updateLocalDeps(absRoot string, gitRoot string, opts updateLocalDepsOptions) (map[string]ModuleAnnotation, error) {
	annotations := make(map[string]ModuleAnnotation)
	processed := make(map[string]bool)
	state := newUpdateLocalDepsState()

	for {
		modules, err := Find(absRoot)
		if err != nil {
			return nil, err
		}
		if len(processed) == len(modules) {
			break
		}

		moduleByDir := make(map[string]Module, len(modules))
		for _, module := range modules {
			moduleByDir[module.Dir] = module
		}

		var ready []Module
		for _, module := range modules {
			if processed[module.Dir] {
				continue
			}
			if moduleDepsProcessed(module, processed) {
				ready = append(ready, module)
			}
		}
		if len(ready) == 0 {
			return nil, fmt.Errorf("local module dependency cycle detected among: %s", remainingModuleDirs(modules, processed))
		}

		for _, module := range ready {
			annotation, err := processModuleLocalDeps(absRoot, gitRoot, module, moduleByDir, state, opts)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", depGoModPath(module.Dir), err)
			}
			if hasAnnotation(annotation) {
				annotations[module.Dir] = mergeAnnotation(annotations[module.Dir], annotation)
			}
			processed[module.Dir] = true
		}
	}

	return annotations, nil
}

func moduleDepsProcessed(module Module, processed map[string]bool) bool {
	for _, dep := range module.Depends {
		if !processed[dep] {
			return false
		}
	}
	return true
}

func remainingModuleDirs(modules []Module, processed map[string]bool) string {
	var dirs []string
	for _, module := range modules {
		if !processed[module.Dir] {
			dirs = append(dirs, depGoModPath(module.Dir))
		}
	}
	return strings.Join(dirs, ", ")
}

func processModuleLocalDeps(absRoot string, gitRoot string, module Module, moduleByDir map[string]Module, state *updateLocalDepsState, opts updateLocalDepsOptions) (ModuleAnnotation, error) {
	moduleAbsDir := absModuleDir(absRoot, module.Dir)
	if opts.DryRun {
		updatedDeps, err := collectModuleLocalDepUpdates(absRoot, module, moduleByDir, state)
		if err != nil {
			return ModuleAnnotation{}, err
		}
		newTag, err := planTagModuleHead(moduleAbsDir, module.Dir, len(updatedDeps) > 0, state)
		if err != nil {
			return ModuleAnnotation{}, err
		}
		return ModuleAnnotation{
			UpdatedDeps: updatedDeps,
			NewTag:      newTag,
		}, nil
	}

	updatedDeps, err := updateModuleGoModLocalDeps(absRoot, moduleAbsDir, module, moduleByDir, state)
	if err != nil {
		return ModuleAnnotation{}, err
	}

	annotation := ModuleAnnotation{}
	if len(updatedDeps) > 0 {
		if err := goModTidy(moduleAbsDir); err != nil {
			return ModuleAnnotation{}, err
		}
		annotation.UpdatedDeps = updatedDeps
		if err := commitModuleChanges(gitRoot, moduleAbsDir, updatedDeps); err != nil {
			return ModuleAnnotation{}, err
		}
	}

	newTag, err := tagModuleHead(moduleAbsDir)
	if err != nil {
		return ModuleAnnotation{}, err
	}
	if newTag != "" {
		state.plannedTags[module.Dir] = newTag
	}
	annotation.NewTag = newTag
	return annotation, nil
}

func updateModuleGoModLocalDeps(absRoot string, moduleAbsDir string, module Module, moduleByDir map[string]Module, state *updateLocalDepsState) ([]DependencyAnnotation, error) {
	updatedDeps, err := collectModuleLocalDepUpdates(absRoot, module, moduleByDir, state)
	if err != nil {
		return nil, err
	}

	for _, dep := range updatedDeps {
		if dep.OldVersion != dep.NewVersion {
			if err := goModEditRequire(moduleAbsDir, dep.ModulePath, dep.NewVersion); err != nil {
				return nil, err
			}
		}
		if dep.RemovedReplace {
			if err := goModEditDropReplace(moduleAbsDir, dep.ModulePath); err != nil {
				return nil, err
			}
		}
	}

	return updatedDeps, nil
}

func collectModuleLocalDepUpdates(absRoot string, module Module, moduleByDir map[string]Module, state *updateLocalDepsState) ([]DependencyAnnotation, error) {
	requireVersions := make(map[string]string, len(module.Requires))
	for _, req := range module.Requires {
		requireVersions[req.Path] = req.Version
	}

	replaced := make(map[string]bool, len(module.Replaces))
	for _, repl := range module.Replaces {
		replaced[repl.OldPath] = true
	}

	var updatedDeps []DependencyAnnotation
	for _, depDir := range module.Depends {
		dep, ok := moduleByDir[depDir]
		if !ok {
			return nil, fmt.Errorf("local dependency %s was not found", depGoModPath(depDir))
		}
		if dep.Path == "" {
			return nil, fmt.Errorf("local dependency %s has empty module path", depGoModPath(depDir))
		}

		latestVersion, err := latestModuleVersion(absRoot, dep, state)
		if err != nil {
			return nil, err
		}

		currentVersion := requireVersions[dep.Path]
		removeReplace := replaced[dep.Path]
		if currentVersion == latestVersion && !removeReplace {
			continue
		}

		updatedDeps = append(updatedDeps, DependencyAnnotation{
			Dir:            dep.Dir,
			ModulePath:     dep.Path,
			OldVersion:     currentVersion,
			NewVersion:     latestVersion,
			RemovedReplace: removeReplace,
		})
	}

	return updatedDeps, nil
}

func latestModuleVersion(absRoot string, dep Module, state *updateLocalDepsState) (string, error) {
	if state != nil {
		if plannedTag := state.plannedTags[dep.Dir]; plannedTag != "" {
			versionPrefix, err := tag.GetVersionPrefix(absModuleDir(absRoot, dep.Dir))
			if err != nil {
				return "", err
			}
			return tag.StripVersionPrefix(versionPrefix, plannedTag), nil
		}
	}

	depAbsDir := absModuleDir(absRoot, dep.Dir)
	versionPrefix, err := tag.GetVersionPrefix(depAbsDir)
	if err != nil {
		return "", err
	}
	latestTag, err := tag.GetLatestVersionTag(depAbsDir, versionPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to get latest tag for %s: %w", depGoModPath(dep.Dir), err)
	}
	return tag.StripVersionPrefix(versionPrefix, latestTag), nil
}

func planTagModuleHead(moduleAbsDir string, moduleDir string, plannedCommit bool, state *updateLocalDepsState) (string, error) {
	versionPrefix, err := tag.GetVersionPrefix(moduleAbsDir)
	if err != nil {
		return "", err
	}

	if !plannedCommit {
		if _, err := tag.GetVersionTag(moduleAbsDir, "HEAD", versionPrefix); err == nil {
			return "", nil
		} else if !errors.Is(err, tag.ErrNoTag) {
			return "", err
		}
	}

	nextTag, err := nextModuleTag(moduleAbsDir, versionPrefix)
	if err != nil {
		return "", err
	}
	state.plannedTags[moduleDir] = nextTag
	return nextTag, nil
}

func tagModuleHead(moduleAbsDir string) (string, error) {
	versionPrefix, err := tag.GetVersionPrefix(moduleAbsDir)
	if err != nil {
		return "", err
	}

	if _, err := tag.GetVersionTag(moduleAbsDir, "HEAD", versionPrefix); err == nil {
		return "", nil
	} else if !errors.Is(err, tag.ErrNoTag) {
		return "", err
	}

	nextTag, err := nextModuleTag(moduleAbsDir, versionPrefix)
	if err != nil {
		return "", err
	}
	if err := runGitCmd(moduleAbsDir, "tag", nextTag); err != nil {
		return "", err
	}
	if err := runGitCmd(moduleAbsDir, "push", "origin", nextTag); err != nil {
		return "", err
	}
	return nextTag, nil
}

func nextModuleTag(moduleAbsDir string, versionPrefix string) (string, error) {
	latestTag, err := tag.GetLatestVersionTag(moduleAbsDir, versionPrefix)
	if err != nil {
		if errors.Is(err, tag.ErrNoTag) {
			return initialModuleTag(versionPrefix), nil
		}
		return "", err
	}
	return git_tag_next.IncrementTag(latestTag)
}

func initialModuleTag(versionPrefix string) string {
	if versionPrefix == "" {
		return "v0.0.1"
	}
	return versionPrefix + "0.0.1"
}

func commitModuleChanges(gitRoot string, moduleAbsDir string, updatedDeps []DependencyAnnotation) error {
	goModPath := filepath.Join(moduleAbsDir, "go.mod")
	relGoModPath, err := filepath.Rel(gitRoot, goModPath)
	if err != nil {
		return err
	}
	relGoModPath = filepath.ToSlash(relGoModPath)
	if err := runGitCmd(gitRoot, "add", "--", relGoModPath); err != nil {
		return err
	}

	goSumPath := filepath.Join(moduleAbsDir, "go.sum")
	relGoSumPath, err := filepath.Rel(gitRoot, goSumPath)
	if err != nil {
		return err
	}
	relGoSumPath = filepath.ToSlash(relGoSumPath)
	if pathExists(goSumPath) || gitPathTracked(gitRoot, relGoSumPath) {
		if err := runGitCmd(gitRoot, "add", "-A", "--", relGoSumPath); err != nil {
			return err
		}
	}
	return runGitCmd(gitRoot, "commit", "-m", localDepsCommitMessage(updatedDeps))
}

func localDepsCommitMessage(updatedDeps []DependencyAnnotation) string {
	parts := make([]string, 0, len(updatedDeps))
	for _, dep := range updatedDeps {
		name := dep.Dir
		if name == "." || name == "" {
			name = dep.ModulePath
		}
		parts = append(parts, fmt.Sprintf("%s to %s", name, dep.NewVersion))
	}
	return "upgrade " + strings.Join(parts, ", ")
}

func absModuleDir(absRoot string, dir string) string {
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	if dir == "." || dir == "" {
		return absRoot
	}
	return filepath.Join(absRoot, filepath.FromSlash(dir))
}

func requireGitRepo(dir string) (string, error) {
	gitRoot, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("%s is not inside a git repo: %w", dir, err)
	}
	gitRoot = strings.TrimSpace(gitRoot)
	if gitRoot == "" {
		return "", fmt.Errorf("%s is not inside a git repo", dir)
	}
	gitRoot, err = filepath.EvalSymlinks(gitRoot)
	if err != nil {
		return "", err
	}
	return gitRoot, nil
}

func requireCleanGitRepo(gitRoot string) error {
	status, err := gitOutput(gitRoot, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) != "" {
		return fmt.Errorf("git repo must be clean before --update-local-deps:\n%s", strings.TrimRight(status, "\n"))
	}
	return nil
}

func goModEditRequire(dir string, modulePath string, version string) error {
	return runCmd(dir, "go", "mod", "edit", "-require="+modulePath+"@"+version)
}

func goModEditDropReplace(dir string, modulePath string) error {
	return runCmd(dir, "go", "mod", "edit", "-dropreplace="+modulePath)
}

func goModTidy(dir string) error {
	return runCmd(dir, "go", "mod", "tidy")
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", formatCmdError("git", args, err, output)
	}
	return string(output), nil
}

func gitPathTracked(dir string, relPath string) bool {
	cmd := exec.Command("git", "ls-files", "--error-unmatch", "--", relPath)
	cmd.Dir = dir
	return cmd.Run() == nil
}

func runGitCmd(dir string, args ...string) error {
	return runCmd(dir, "git", args...)
}

func runCmd(dir string, name string, args ...string) error {
	fmt.Fprintln(os.Stderr, name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func formatCmdError(name string, args []string, err error, output []byte) error {
	output = bytes.TrimSpace(output)
	if len(output) == 0 {
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, output)
}

func hasAnnotation(annotation ModuleAnnotation) bool {
	return len(annotation.UpdatedDeps) > 0 || annotation.NewTag != ""
}

func mergeAnnotation(a ModuleAnnotation, b ModuleAnnotation) ModuleAnnotation {
	a.UpdatedDeps = append(a.UpdatedDeps, b.UpdatedDeps...)
	if b.NewTag != "" {
		a.NewTag = b.NewTag
	}
	return a
}
