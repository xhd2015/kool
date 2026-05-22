package modules

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFindSkipsVendorAndGitIgnoredDirs(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "go.mod"), `module example.com/root

require (
	example.com/app v0.0.0
	example.com/service v0.0.0
)
`)
	mustWrite(t, filepath.Join(dir, "app", "go.mod"), `module example.com/app

require example.com/service v0.0.0
`)
	mustWrite(t, filepath.Join(dir, "nested", "service", "go.mod"), "module example.com/service\n")
	mustWrite(t, filepath.Join(dir, "template", "go.mod"), "module example.com/{{.Name}}\n\ngo {{.GoVersion}}\n")
	mustWrite(t, filepath.Join(dir, "vendor", "ignored", "go.mod"), "module example.com/vendor\n")
	mustWrite(t, filepath.Join(dir, "ignored", "go.mod"), "module example.com/ignored\n")
	mustWrite(t, filepath.Join(dir, "nested-ignore", ".gitignore"), ".xgo/gen/\n")
	mustWrite(t, filepath.Join(dir, "nested-ignore", ".xgo", "gen", "go.mod"), "module example.com/xgo\n")
	mustWrite(t, filepath.Join(dir, ".gitignore"), "ignored/\n")

	mustRunGit(t, dir, "init")

	modules, err := Find(dir)
	if err != nil {
		t.Fatal(err)
	}

	got := moduleSummaries(modules)
	want := []moduleSummary{
		{Dir: ".", Path: "example.com/root", Depends: []string{"app", "nested/service"}},
		{Dir: "app", Path: "example.com/app", Depends: []string{"nested/service"}},
		{Dir: "nested/service", Path: "example.com/service"},
		{Dir: "template", Path: "example.com/{{.Name}}"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("modules mismatch\nwant: %#v\n got: %#v", want, got)
	}
}

func TestRender(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, []Module{
		{
			Dir:            ".",
			Path:           "example.com/root",
			Depends:        []string{"app", "nested/service"},
			LatestTag:      "v0.0.3",
			LatestTagKnown: true,
			Requires: []ModuleRequire{
				{Path: "example.com/app", Version: "v0.0.2"},
				{Path: "example.com/service", Version: "v0.0.4"},
			},
		},
		{
			Dir:            "app",
			Path:           "example.com/app",
			Depends:        []string{"nested/service"},
			LatestTag:      "app/v0.0.2",
			LatestTagKnown: true,
			Requires: []ModuleRequire{
				{Path: "example.com/service", Version: "v0.0.4"},
			},
		},
		{
			Dir:            "nested/service",
			Path:           "example.com/service",
			LatestTag:      "nested/service/v0.0.4",
			LatestTagKnown: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := strings.TrimLeft(`
.
├── app
│   └── go.mod [latest tag: app/v0.0.2]
│       └── (depends on) nested/service/go.mod [version: v0.0.4]
├── go.mod [latest tag: v0.0.3]
│   ├── (depends on) app/go.mod [version: v0.0.2]
│   └── (depends on) nested/service/go.mod [version: v0.0.4]
└── nested
    └── service
        └── go.mod [latest tag: nested/service/v0.0.4]
`, "\n")
	if buf.String() != want {
		t.Fatalf("tree mismatch\nwant:\n%s\ngot:\n%s", want, buf.String())
	}
}

func TestFindPopulatesLatestTags(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "go.mod"), `module example.com/root

require example.com/app v0.0.2
`)
	mustWrite(t, filepath.Join(dir, "app", "go.mod"), "module example.com/app\n")
	mustRunGit(t, dir, "init")
	mustRunGit(t, dir, "config", "user.email", "test@example.com")
	mustRunGit(t, dir, "config", "user.name", "Test User")
	mustRunGit(t, dir, "add", ".")
	mustRunGit(t, dir, "commit", "-m", "initial")
	mustRunGit(t, dir, "tag", "v0.0.1")
	mustRunGit(t, dir, "tag", "v0.0.2")
	mustRunGit(t, dir, "tag", "app/v0.0.3")

	modules, err := Find(dir)
	if err != nil {
		t.Fatal(err)
	}

	byDir := make(map[string]Module, len(modules))
	for _, module := range modules {
		byDir[module.Dir] = module
	}
	if got := byDir["."].LatestTag; got != "v0.0.2" {
		t.Fatalf("root latest tag mismatch: %q", got)
	}
	if got := byDir["app"].LatestTag; got != "app/v0.0.3" {
		t.Fatalf("app latest tag mismatch: %q", got)
	}

	var buf bytes.Buffer
	if err := Render(&buf, modules); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"go.mod [latest tag: v0.0.2]",
		"(depends on) app/go.mod [version: v0.0.2]",
		"go.mod [latest tag: app/v0.0.3]",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("rendered tree missing %q:\n%s", want, buf.String())
		}
	}

	var noTagsBuf bytes.Buffer
	if err := handle(&noTagsBuf, []string{"--dir", dir, "--no-tags"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(noTagsBuf.String(), "latest tag:") {
		t.Fatalf("--no-tags output should not include latest tag annotations:\n%s", noTagsBuf.String())
	}
	if !strings.Contains(noTagsBuf.String(), "(depends on) app/go.mod [version: v0.0.2]") {
		t.Fatalf("--no-tags output should keep dependency versions:\n%s", noTagsBuf.String())
	}
}

func TestListModuleFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".gitignore"), "*.log\nignored.tmp\n")
	mustWrite(t, filepath.Join(dir, "go.mod"), "module example.com/root\n")
	mustWrite(t, filepath.Join(dir, "root.go"), "package root\n")
	mustWrite(t, filepath.Join(dir, "types", "go.mod"), "module example.com/root/types\n")
	mustWrite(t, filepath.Join(dir, "types", "types.go"), "package types\n")
	mustWrite(t, filepath.Join(dir, "types", "inner", "go.mod"), "module example.com/root/types/inner\n")
	mustWrite(t, filepath.Join(dir, "types", "inner", "inner.go"), "package inner\n")
	mustRunGit(t, dir, "init")
	mustRunGit(t, dir, "config", "user.email", "test@example.com")
	mustRunGit(t, dir, "config", "user.name", "Test User")
	mustRunGit(t, dir, "add", ".")
	mustRunGit(t, dir, "commit", "-m", "initial")
	mustWrite(t, filepath.Join(dir, "root-untracked.txt"), "root untracked\n")
	mustWrite(t, filepath.Join(dir, "types", "untracked.txt"), "types untracked\n")
	mustWrite(t, filepath.Join(dir, "types", "ignored.log"), "ignored\n")
	mustWrite(t, filepath.Join(dir, "ignored.tmp"), "ignored\n")
	mustWrite(t, filepath.Join(dir, "types", "inner", "untracked.txt"), "inner untracked\n")

	rootFiles, err := ListModuleFiles(dir, ".")
	if err != nil {
		t.Fatal(err)
	}
	wantRoot := []string{".gitignore", "go.mod", "root-untracked.txt", "root.go"}
	if !reflect.DeepEqual(rootFiles, wantRoot) {
		t.Fatalf("root files mismatch\nwant: %#v\n got: %#v", wantRoot, rootFiles)
	}

	typesFiles, err := ListModuleFiles(dir, "types")
	if err != nil {
		t.Fatal(err)
	}
	wantTypes := []string{"types/go.mod", "types/types.go", "types/untracked.txt"}
	if !reflect.DeepEqual(typesFiles, wantTypes) {
		t.Fatalf("types files mismatch\nwant: %#v\n got: %#v", wantTypes, typesFiles)
	}

	var buf bytes.Buffer
	if err := handle(&buf, []string{"--dir", dir, "ls-files", "--module", "types"}); err != nil {
		t.Fatal(err)
	}
	wantOutput := strings.Join(wantTypes, "\n") + "\n"
	if buf.String() != wantOutput {
		t.Fatalf("ls-files output mismatch\nwant:\n%s\ngot:\n%s", wantOutput, buf.String())
	}
}

func TestUpdateLocalDepsOldFlagsAreRejected(t *testing.T) {
	var buf bytes.Buffer
	err := handle(&buf, []string{"--update-local-deps"})
	if err == nil || !strings.Contains(err.Error(), "unrecognized flag: --update-local-deps") {
		t.Fatalf("expected --update-local-deps to be rejected, got %v", err)
	}
	err = handle(&buf, []string{"--dry-run"})
	if err == nil || !strings.Contains(err.Error(), "unrecognized flag: --dry-run") {
		t.Fatalf("expected top-level --dry-run to be rejected, got %v", err)
	}
	err = handle(&buf, []string{"update-local-deps", "--dry-run", "--no-tags"})
	if err == nil || !strings.Contains(err.Error(), "unrecognized flag: --no-tags") {
		t.Fatalf("expected update-local-deps --no-tags to be rejected, got %v", err)
	}
	err = handle(&buf, []string{"--no-tags", "update-local-deps", "--dry-run"})
	if err == nil || !strings.Contains(err.Error(), "--no-tags is not supported with update-local-deps") {
		t.Fatalf("expected leading --no-tags update-local-deps to be rejected, got %v", err)
	}
}

func TestUpdateLocalDepsAndRender(t *testing.T) {
	dir, origin := setupLocalDepsRepo(t)

	var buf bytes.Buffer
	if err := UpdateLocalDepsAndRender(&buf, dir, false); err != nil {
		t.Fatal(err)
	}

	rootGoMod := mustRead(t, filepath.Join(dir, "go.mod"))
	for _, want := range []string{
		"example.com/root.git/cli v0.0.1",
		"example.com/root.git/types v0.0.1",
	} {
		if !strings.Contains(rootGoMod, want) {
			t.Fatalf("root go.mod missing %q:\n%s", want, rootGoMod)
		}
	}
	if strings.Contains(rootGoMod, "replace example.com/root.git/") {
		t.Fatalf("root go.mod still has local replace:\n%s", rootGoMod)
	}
	if _, err := os.Stat(filepath.Join(dir, "go.sum")); err != nil {
		t.Fatalf("root go.sum was not generated by go mod tidy: %v", err)
	}

	cliGoMod := mustRead(t, filepath.Join(dir, "cli", "go.mod"))
	if !strings.Contains(cliGoMod, "example.com/root.git/types v0.0.1") {
		t.Fatalf("cli go.mod was not updated:\n%s", cliGoMod)
	}
	if strings.Contains(cliGoMod, "replace example.com/root.git/types") {
		t.Fatalf("cli go.mod still has local replace:\n%s", cliGoMod)
	}
	if _, err := os.Stat(filepath.Join(dir, "cli", "go.sum")); err != nil {
		t.Fatalf("cli go.sum was not generated by go mod tidy: %v", err)
	}

	for _, tag := range []string{"types/v0.0.1", "cli/v0.0.1", "v0.0.1"} {
		mustRunGit(t, origin, "rev-parse", "--verify", "refs/tags/"+tag)
		if !strings.Contains(buf.String(), "new tag: "+tag) {
			t.Fatalf("rendered tree missing annotation for %s:\n%s", tag, buf.String())
		}
	}
	if status := mustRunGitOutput(t, dir, "status", "--short"); status != "" {
		t.Fatalf("update left uncommitted changes:\n%s", status)
	}
	if !strings.Contains(buf.String(), "updated: types/go.mod v0.0.0 -> v0.0.1, replace removed") {
		t.Fatalf("rendered tree missing cli dependency annotation:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "updated: cli/go.mod v0.0.0 -> v0.0.1, replace removed") {
		t.Fatalf("rendered tree missing root dependency annotation:\n%s", buf.String())
	}
}

func TestUpdateLocalDepsAndRenderDryRun(t *testing.T) {
	dir, origin := setupLocalDepsRepo(t)

	rootGoModBefore := mustRead(t, filepath.Join(dir, "go.mod"))
	cliGoModBefore := mustRead(t, filepath.Join(dir, "cli", "go.mod"))
	headBefore := mustRunGitOutput(t, dir, "rev-parse", "HEAD")

	var buf bytes.Buffer
	if err := handle(&buf, []string{"--dir", dir, "update-local-deps", "--dry-run"}); err != nil {
		t.Fatal(err)
	}

	if rootGoMod := mustRead(t, filepath.Join(dir, "go.mod")); rootGoMod != rootGoModBefore {
		t.Fatalf("dry-run changed root go.mod\nbefore:\n%s\nafter:\n%s", rootGoModBefore, rootGoMod)
	}
	if cliGoMod := mustRead(t, filepath.Join(dir, "cli", "go.mod")); cliGoMod != cliGoModBefore {
		t.Fatalf("dry-run changed cli go.mod\nbefore:\n%s\nafter:\n%s", cliGoModBefore, cliGoMod)
	}
	if headAfter := mustRunGitOutput(t, dir, "rev-parse", "HEAD"); headAfter != headBefore {
		t.Fatalf("dry-run changed HEAD: before %s after %s", headBefore, headAfter)
	}
	if status := mustRunGitOutput(t, dir, "status", "--short"); status != "" {
		t.Fatalf("dry-run changed git status:\n%s", status)
	}
	if tags := mustRunGitOutput(t, dir, "tag", "-l"); tags != "" {
		t.Fatalf("dry-run created local tags:\n%s", tags)
	}
	if tags := mustRunGitOutput(t, origin, "tag", "-l"); tags != "" {
		t.Fatalf("dry-run pushed tags:\n%s", tags)
	}

	for _, want := range []string{
		"new tag: types/v0.0.1",
		"updated: types/go.mod v0.0.0 -> v0.0.1, replace removed",
		"new tag: cli/v0.0.1",
		"updated: cli/go.mod v0.0.0 -> v0.0.1, replace removed",
		"new tag: v0.0.1",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, buf.String())
		}
	}

}

func TestUpdateLocalDepsSkipsTagForUnchangedSubmoduleTree(t *testing.T) {
	dir, origin := setupTaggedRootReadmeRepo(t)

	var buf bytes.Buffer
	if err := UpdateLocalDepsAndRender(&buf, dir, false); err != nil {
		t.Fatal(err)
	}

	if gitTagExists(origin, "types/v0.0.2") {
		t.Fatalf("unchanged types tree should not create types/v0.0.2")
	}
	if !gitTagExists(origin, "v0.0.2") {
		t.Fatalf("changed root tree should create and push v0.0.2")
	}
	if strings.Contains(buf.String(), "new tag: types/v0.0.1 -> types/v0.0.2") || strings.Contains(buf.String(), "new tag: types/v0.0.2") {
		t.Fatalf("rendered tree should not annotate unchanged types tag:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "new tag: v0.0.1 -> v0.0.2") {
		t.Fatalf("rendered tree missing root tag annotation:\n%s", buf.String())
	}
	if status := mustRunGitOutput(t, dir, "status", "--short"); status != "" {
		t.Fatalf("update left uncommitted changes:\n%s", status)
	}
}

func TestUpdateLocalDepsSkipsParentTagForNestedModuleChange(t *testing.T) {
	dir, origin := setupNestedModuleChangeRepo(t)

	var buf bytes.Buffer
	if err := UpdateLocalDepsAndRender(&buf, dir, false); err != nil {
		t.Fatal(err)
	}

	if gitTagExists(origin, "v0.0.2") {
		t.Fatalf("parent module should not create v0.0.2 for nested module-only change")
	}
	if !gitTagExists(origin, "inner/v0.0.2") {
		t.Fatalf("changed nested module should create and push inner/v0.0.2")
	}
	if strings.Contains(buf.String(), "new tag: v0.0.1 -> v0.0.2") || strings.Contains(buf.String(), "new tag: v0.0.2") {
		t.Fatalf("rendered tree should not annotate parent tag for nested module-only change:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "new tag: inner/v0.0.1 -> inner/v0.0.2") {
		t.Fatalf("rendered tree missing nested module tag annotation:\n%s", buf.String())
	}
	if status := mustRunGitOutput(t, dir, "status", "--short"); status != "" {
		t.Fatalf("update left uncommitted changes:\n%s", status)
	}
}

func setupLocalDepsRepo(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	origin := filepath.Join(t.TempDir(), "origin.git")

	mustRunGit(t, dir, "init")
	mustRunGit(t, dir, "config", "user.email", "test@example.com")
	mustRunGit(t, dir, "config", "user.name", "Test User")
	if err := os.MkdirAll(origin, 0755); err != nil {
		t.Fatal(err)
	}
	mustRunGit(t, origin, "init", "--bare")
	mustRunGit(t, dir, "remote", "add", "origin", origin)
	configGoFetchFromOrigin(t, origin)

	mustWrite(t, filepath.Join(dir, "go.mod"), `module example.com/root.git
go 1.23.0

require (
	example.com/root.git/cli v0.0.0
	example.com/root.git/types v0.0.0
)

replace example.com/root.git/cli => ./cli

replace example.com/root.git/types => ./types
`)
	mustWrite(t, filepath.Join(dir, "root.go"), `package root

import (
	_ "example.com/root.git/cli"
	_ "example.com/root.git/types"
)
`)
	mustWrite(t, filepath.Join(dir, "types", "go.mod"), `module example.com/root.git/types
go 1.23.0
`)
	mustWrite(t, filepath.Join(dir, "types", "types.go"), "package types\n")
	mustWrite(t, filepath.Join(dir, "cli", "go.mod"), `module example.com/root.git/cli
go 1.23.0

require example.com/root.git/types v0.0.0

replace example.com/root.git/types => ../types
`)
	mustWrite(t, filepath.Join(dir, "cli", "cli.go"), `package cli

import _ "example.com/root.git/types"
`)
	mustRunGit(t, dir, "add", ".")
	mustRunGit(t, dir, "commit", "-m", "initial")
	return dir, origin
}

func setupTaggedRootReadmeRepo(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	origin := filepath.Join(t.TempDir(), "origin.git")

	mustRunGit(t, dir, "init")
	mustRunGit(t, dir, "config", "user.email", "test@example.com")
	mustRunGit(t, dir, "config", "user.name", "Test User")
	if err := os.MkdirAll(origin, 0755); err != nil {
		t.Fatal(err)
	}
	mustRunGit(t, origin, "init", "--bare")
	mustRunGit(t, dir, "remote", "add", "origin", origin)

	mustWrite(t, filepath.Join(dir, "go.mod"), `module example.com/root.git
go 1.23.0

require example.com/root.git/types v0.0.1
`)
	mustWrite(t, filepath.Join(dir, "root.go"), `package root

import _ "example.com/root.git/types"
`)
	mustWrite(t, filepath.Join(dir, "types", "go.mod"), `module example.com/root.git/types
go 1.23.0
`)
	mustWrite(t, filepath.Join(dir, "types", "types.go"), "package types\n")
	mustRunGit(t, dir, "add", ".")
	mustRunGit(t, dir, "commit", "-m", "initial")
	mustRunGit(t, dir, "tag", "v0.0.1")
	mustRunGit(t, dir, "tag", "types/v0.0.1")
	mustRunGit(t, dir, "push", "origin", "v0.0.1", "types/v0.0.1")
	mustWrite(t, filepath.Join(dir, "README.md"), "update README\n")
	mustRunGit(t, dir, "add", "README.md")
	mustRunGit(t, dir, "commit", "-m", "update README")
	return dir, origin
}

func setupNestedModuleChangeRepo(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	origin := filepath.Join(t.TempDir(), "origin.git")

	mustRunGit(t, dir, "init")
	mustRunGit(t, dir, "config", "user.email", "test@example.com")
	mustRunGit(t, dir, "config", "user.name", "Test User")
	if err := os.MkdirAll(origin, 0755); err != nil {
		t.Fatal(err)
	}
	mustRunGit(t, origin, "init", "--bare")
	mustRunGit(t, dir, "remote", "add", "origin", origin)

	mustWrite(t, filepath.Join(dir, "go.mod"), `module example.com/root.git
go 1.23.0
`)
	mustWrite(t, filepath.Join(dir, "root.go"), "package root\n")
	mustWrite(t, filepath.Join(dir, "inner", "go.mod"), `module example.com/root.git/inner
go 1.23.0
`)
	mustWrite(t, filepath.Join(dir, "inner", "inner.go"), "package inner\n")
	mustRunGit(t, dir, "add", ".")
	mustRunGit(t, dir, "commit", "-m", "initial")
	mustRunGit(t, dir, "tag", "v0.0.1")
	mustRunGit(t, dir, "tag", "inner/v0.0.1")
	mustRunGit(t, dir, "push", "origin", "v0.0.1", "inner/v0.0.1")
	mustWrite(t, filepath.Join(dir, "inner", "inner.go"), "package inner\n\nconst Version = 2\n")
	mustRunGit(t, dir, "add", ".")
	mustRunGit(t, dir, "commit", "-m", "update inner")
	return dir, origin
}

func configGoFetchFromOrigin(t *testing.T, origin string) {
	t.Helper()
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	originURL := "file://" + filepath.ToSlash(origin)
	mustWrite(t, gitConfig, `[url "`+originURL+`"]
	insteadOf = https://example.com/root.git
	insteadOf = http://example.com/root.git
	insteadOf = git://example.com/root.git
	insteadOf = ssh://example.com/root.git
	insteadOf = git+ssh://example.com/root.git
	insteadOf = https://example.com/root
	insteadOf = http://example.com/root
	insteadOf = git://example.com/root
	insteadOf = ssh://example.com/root
	insteadOf = git+ssh://example.com/root
`)
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)
	t.Setenv("GIT_ALLOW_PROTOCOL", "file:https:http:git:ssh")
	t.Setenv("GOPRIVATE", "example.com")
	t.Setenv("GONOSUMDB", "example.com")
	t.Setenv("GOPROXY", "direct")
	t.Setenv("GOINSECURE", "example.com")
}

type moduleSummary struct {
	Dir     string
	Path    string
	Depends []string
}

func moduleSummaries(modules []Module) []moduleSummary {
	summaries := make([]moduleSummary, 0, len(modules))
	for _, module := range modules {
		summaries = append(summaries, moduleSummary{
			Dir:     module.Dir,
			Path:    module.Path,
			Depends: module.Depends,
		})
	}
	return summaries
}

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func mustRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func mustRunGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func gitTagExists(dir string, tag string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", "refs/tags/"+tag)
	cmd.Dir = dir
	return cmd.Run() == nil
}
