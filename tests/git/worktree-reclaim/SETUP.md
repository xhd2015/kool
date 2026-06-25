# Scenario

**Feature**: kool git worktree reclaim removes clean linked worktrees whose HEAD is included in main

```
# user invokes reclaim; handler resolves main repo and evaluates candidates
user -> kool git worktree reclaim <dir>|--all [--dry-run] -> reclaim handler

# git checks cleanliness and HEAD inclusion; handler removes worktree + branch
reclaim handler -> git worktree list/status/compare -> git worktree remove + branch -D
```

## Preconditions

- The `kool` command is available in PATH
- Git is available in PATH

## Steps

1. Verify `kool` and `git` are available
2. Execute `kool git worktree reclaim` with configured Request fields
3. Capture stdout, stderr, and exit code

## Context

- Reclaim is conservative and non-interactive: only clean worktrees whose HEAD is an ancestor of (or equal to) main `HEAD` are removed
- Single-path reclaim fails with non-zero exit when the target is not reclaimable
- `--all` skips non-reclaimable worktrees but exits non-zero only on removal errors

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
		return fmt.Errorf("kool not found in PATH, build it first: %w", err)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}
	return nil
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s failed: %v\n%s", args, dir, err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func initMainRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mainRepo := filepath.Join(dir, "main")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		t.Fatalf("mkdir main repo: %v", err)
	}
	runGit(t, mainRepo, "init")
	runGit(t, mainRepo, "config", "user.email", "test@test.com")
	runGit(t, mainRepo, "config", "user.name", "test")
	writeFile(t, filepath.Join(mainRepo, "README.md"), "# main\n")
	runGit(t, mainRepo, "add", ".")
	runGit(t, mainRepo, "commit", "-m", "initial commit")
	runGit(t, mainRepo, "branch", "-M", "main")
	return mainRepo
}

func addLinkedWorktree(t *testing.T, mainRepo, wtName, branch string) string {
	t.Helper()
	wtPath := filepath.Join(filepath.Dir(mainRepo), wtName)
	runGit(t, mainRepo, "worktree", "add", "-b", branch, wtPath)
	return wtPath
}

func mergeBranch(t *testing.T, mainRepo, branch string) {
	t.Helper()
	runGit(t, mainRepo, "merge", branch)
}

func pathExists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	return err == nil
}

func branchExists(t *testing.T, mainRepo, branch string) bool {
	t.Helper()
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = mainRepo
	return cmd.Run() == nil
}

func deleteWorktreeDir(t *testing.T, wtPath string) {
	t.Helper()
	if err := os.RemoveAll(wtPath); err != nil {
		t.Fatalf("remove worktree dir %s: %v", wtPath, err)
	}
}

func worktreeListed(t *testing.T, mainRepo, wtPath string) bool {
	t.Helper()
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = mainRepo
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git worktree list in %s: %v\n%s", mainRepo, err, out)
	}
	return strings.Contains(string(out), wtPath)
}

func combinedOutput(resp *Response) string {
	return resp.Stdout + resp.Stderr
}
```