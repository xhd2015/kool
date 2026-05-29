package create

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateGoCLIIntoEmptyExistingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "git-hooks")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := HandleCreateGoCLI([]string{dir}); err != nil {
		t.Fatal(err)
	}

	for _, file := range []string{
		".gitignore",
		"go.mod",
		"go.sum",
		"main.go",
		filepath.Join("run", "run.go"),
	} {
		if _, err := os.Stat(filepath.Join(dir, file)); err != nil {
			t.Fatalf("expected %s: %v", file, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf("expected git repository to be initialized: %v", err)
	}

	mainGo := mustReadCreateTest(t, filepath.Join(dir, "main.go"))
	if !strings.Contains(mainGo, `"git-hooks/run"`) {
		t.Fatalf("main.go did not dispatch to generated run package:\n%s", mainGo)
	}

	runGo := mustReadCreateTest(t, filepath.Join(dir, "run", "run.go"))
	for _, want := range []string{
		"type Config struct{}",
		"func Main(args []string) error",
		"func Run(config Config) error",
		"github.com/xhd2015/less-gen/flags",
	} {
		if !strings.Contains(runGo, want) {
			t.Fatalf("run.go missing %q:\n%s", want, runGo)
		}
	}

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go test ./...: %v\n%s", err, output)
	}
}

func TestCreateGoCLIRejectsDirWithDotFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "git-hooks")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("tmp\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := HandleCreateGoCLI([]string{dir})
	if err == nil {
		t.Fatal("expected non-empty directory error")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("expected not empty error, got: %v", err)
	}
}

func TestCreateGoCLIAllowsDirWithOnlyGit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "git-hooks")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := HandleCreateGoCLI([]string{dir}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf("expected .git to be preserved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		t.Fatalf("expected go.mod to be created: %v", err)
	}
}

func TestPrepareProjectDirAllowsDirWithOnlyGit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myproject")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	created, err := prepareProjectDir(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if created {
		t.Fatal("expected created=false since directory already existed")
	}
}

func TestPrepareProjectDirRejectsDirWithOtherFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myproject")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := prepareProjectDir(dir)
	if err == nil {
		t.Fatal("expected error for non-empty directory")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("expected not empty error, got: %v", err)
	}
}

func TestSuggestGoModPathNoTrailingSlash(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "init")
	runGit(t, dir, "remote", "add", "origin", "ssh://git@github.com/xhd2015/dotfiles-migrate")

	path, err := suggestGoModPath(dir)
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("expected non-empty module path")
	}
	if strings.HasSuffix(path, "/") {
		t.Fatalf("module path must not have trailing slash: %s", path)
	}
	if path != "github.com/xhd2015/dotfiles-migrate" {
		t.Errorf("expected github.com/xhd2015/dotfiles-migrate, got %s", path)
	}
}

func TestCreateGoCLIWithTrailingSlash(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "git-hooks")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := HandleCreateGoCLI([]string{dir + string(filepath.Separator)}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		t.Fatalf("expected go.mod to be created: %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func mustReadCreateTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
