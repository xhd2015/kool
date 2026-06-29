# Scenario

**Feature**: kool go update refreshes a local module dependency to the latest git tag

```
# consumer module with replace -> update target dir -> require latest tag, drop replace
user -> kool go update <dir> -> go_update.Update -> go.mod edited in cwd
```

## Preconditions

- The `kool` command is available in PATH (for CLI leaves)
- `go` and `git` are available in PATH

## Steps

1. Verify tooling is available
2. Build fixture repo and consumer module (per leaf Setup)
3. Run update via library or CLI
4. Assert consumer go.mod reflects the latest tag

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const nestedModulePath = "github.com/example/dot-pkgs/go-pkgs"

func Setup(t *testing.T, req *Request) error {
	_, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	_, err = exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}
	return nil
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func mustGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func initDotPkgsRepo(t *testing.T, workspace string) string {
	t.Helper()
	repo := filepath.Join(workspace, "dot-pkgs")
	goPkgs := filepath.Join(repo, "go-pkgs")

	if err := os.MkdirAll(goPkgs, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, "README.md"), "# dot-pkgs\n")
	writeFile(t, filepath.Join(goPkgs, "go.mod"), "module "+nestedModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(goPkgs, "pkg.go"), "package gopkgs\n")

	mustGit(t, repo, "init")
	mustGit(t, repo, "config", "user.email", "test@example.com")
	mustGit(t, repo, "config", "user.name", "Test User")
	mustGit(t, repo, "add", ".")
	mustGit(t, repo, "commit", "-m", "initial go-pkgs")
	mustGit(t, repo, "tag", "go-pkgs/v0.0.2")

	writeFile(t, filepath.Join(repo, "TODO"), "post-tag change\n")
	mustGit(t, repo, "add", "TODO")
	mustGit(t, repo, "commit", "-m", "post-tag commit")

	return goPkgs
}

func initConsumerModule(t *testing.T, workspace, goPkgsDir string) string {
	t.Helper()
	consumer := filepath.Join(workspace, "consumer")
	if err := os.MkdirAll(consumer, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	mustGo(t, consumer, "mod", "edit", "-require="+nestedModulePath+"@v0.0.1")
	mustGo(t, consumer, "mod", "edit", "-replace="+nestedModulePath+"="+goPkgsDir)
	return consumer
}

func initNestedModuleFixture(t *testing.T, req *Request) {
	t.Helper()
	workspace, err := os.MkdirTemp("", "kool-go-update-nested-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(workspace) })

	goPkgs := initDotPkgsRepo(t, workspace)
	consumer := initConsumerModule(t, workspace, goPkgs)

	req.TargetDir = goPkgs
	req.ConsumerDir = consumer
}
```