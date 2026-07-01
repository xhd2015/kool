# Scenario

**Feature**: kool go modules --list streams `<dir> <module-path>` lines in walk order

```
# user runs kool go modules --list --dir <root>; handler delegates to scan.ScanStream
workspace (go.mod files + git) -> kool go modules --list --dir <root> -> stdout lines

# each line: <dir> <path>; dir is "." for root, slash-relative for sub-dirs (no ./)
scan.ScanStream(root) -> per Module -> fmt.Fprintln(stdout, dir + " " + path)

# skip rules (scan package): .git/vendor/testdata, gitignored, nested separate repos
```

## Preconditions

- The `kool` command is available in PATH, **rebuilt with the local replace directive**
  in `kool-cli/go.mod`:
  ```
  replace github.com/xhd2015/dot-pkgs/go-pkgs => ../go-pkgs
  ```
  so kool picks up the new (unpublished) scan package. Build with
  `go build`/`go install` from `kool-cli/` before running these doctests. Without the
  rebuild, `--list` is not wired up and the leaves fail.
- `go` and `git` are available on PATH.

## Steps

1. Verify `kool`, `go`, and `git` are available.
2. Leaf `Setup` builds an isolated temp workspace: `go.mod` files + git init/config/add/commit
   and/or nested separate repo as the scenario requires.
3. Leaf `Setup` sets `req.RootDir`.
4. Root `Run` execs `kool go modules --list --dir <root>` and captures stdout.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("kool"); err != nil {
		return fmt.Errorf("kool not found in PATH, build it first with the local replace: %w", err)
	}
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}
	return nil
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
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

// newWorkspace returns an isolated temp dir, cleaned up after the test.
func newWorkspace(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "kool-modules-doctest-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// writeModule writes a go.mod declaring modulePath with go 1.22 (hermetic, offline).
func writeModule(t *testing.T, dir, modulePath string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "go.mod"), "module "+modulePath+"\n\ngo 1.22\n")
}

// initGitRepo runs git init + identity config + add + commit in dir.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	mustGit(t, dir, "init", "-b", "main")
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "Test User")
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "init")
}

// stdoutLines splits resp.Stdout into trimmed non-empty lines.
func stdoutLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}
```
