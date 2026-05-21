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

func mustReadCreateTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
