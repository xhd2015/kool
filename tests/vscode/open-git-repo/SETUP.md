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
```