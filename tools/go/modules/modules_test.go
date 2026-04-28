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
		{Dir: ".", Path: "example.com/root", Depends: []string{"app", "nested/service"}},
		{Dir: "app", Path: "example.com/app", Depends: []string{"nested/service"}},
		{Dir: "nested/service", Path: "example.com/service"},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := strings.TrimLeft(`
.
├── app
│   └── go.mod
│       └── (depends on) nested/service/go.mod
├── go.mod
│   ├── (depends on) app/go.mod
│   └── (depends on) nested/service/go.mod
└── nested
    └── service
        └── go.mod
`, "\n")
	if buf.String() != want {
		t.Fatalf("tree mismatch\nwant:\n%s\ngot:\n%s", want, buf.String())
	}
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

func mustRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}
