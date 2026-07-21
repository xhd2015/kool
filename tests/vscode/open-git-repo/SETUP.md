# Scenario

**Feature**: kool opens local git repos in VS Code SCM via vscode:// URI

```
# CLI validates repo and builds deep link
kool vscode open-git-repo <path> -> validateGitRepoPath -> buildGitOpenRepoURI

# OS handler launches/focuses VS Code
buildGitOpenRepoURI -> OS opener (open/xdg-open/cmd) -> VS Code
```

## Preconditions
- Go toolchain available.
- Git available for fixture repos.
- `github.com/xhd2015/kool/vscodegit` package provides testable functions (implementer adds).
- CLI integration tests require built `kool` binary in PATH or module root.

## Steps
1. Each leaf sets `req.Phase` and path fields via `Setup`.
2. `Run` dispatches to validate, URI build, exec mock, or CLI subprocess.
3. Each leaf asserts outcomes via `Assert`.

## Context
- **Extension id**: `xhd2015.open-in-new-window`
- **URI path**: `/git-open?path=<url-encoded-absolute-path>`
- **IPC socket**: `~/.kool/xhd2015.open-in-new-window.sock` (overridable in tests)
- **IPC fallback stderr**: `Note: extension not reachable via IPC; opening via vscode:// URI.`
- **Relative paths**: resolved against cwd before validation

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func expectedExtensionURI(absPath string) string {
	encoded := strings.ReplaceAll(absPath, " ", "%20")
	return fmt.Sprintf("vscode://xhd2015.open-in-new-window/git-open?path=%s", encoded)
}

func Setup(t *testing.T, req *Request) error {
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	if err := os.Chdir(req.WorkingDir); err != nil {
		return err
	}
	return nil
}

func osMkdir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func initGitRepo(t *testing.T, repoDir string) error {
	t.Helper()
	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	runGit("init")
	runGit("config", "user.email", "test@test.com")
	runGit("config", "user.name", "test")
	readme := fmt.Sprintf("%s/README.md", repoDir)
	if err := os.WriteFile(readme, []byte("# test"), 0644); err != nil {
		return err
	}
	runGit("add", ".")
	runGit("commit", "-m", "initial")
	return nil
}

func writeFakeCodeScript(t *testing.T, dir string, listExtensions []string) string {
	t.Helper()
	script := filepath.Join(dir, "code")
	body := `#!/bin/sh
case "$1" in
--list-extensions)
`
	for _, ext := range listExtensions {
		body += fmt.Sprintf("  echo '%s'\n", ext)
	}
	body += `  ;;
*)
  echo "unknown: $1" >&2
  exit 1
  ;;
esac
`
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatalf("write fake code: %v", err)
	}
	return script
}

func installExtensionListedPrecheck(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.WorkingDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	req.CodeCommand = writeFakeCodeScript(t, binDir, []string{"xhd2015.open-in-new-window"})
	req.CodeInPath = true
}

func installNoExtensionPrecheck(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.WorkingDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	req.CodeCommand = writeFakeCodeScript(t, binDir, []string{"other.extension"})
	req.CodeInPath = true
}

func initValidGitRepo(t *testing.T, baseDir string, name string) string {
	t.Helper()
	repoDir := filepath.Join(baseDir, name)
	if err := osMkdir(repoDir); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := initGitRepo(t, repoDir); err != nil {
		t.Fatalf("init git: %v", err)
	}
	return repoDir
}

// markRootTree keeps hierarchical child packages importing this package live.
func markRootTree() {}
```