# Scenario

**Feature**: kool git worktree merge-back merges a linked worktree branch into a target checkout

```
# user invokes merge-back from linked worktree cwd
user -> kool git worktree merge-back [--to <target>] [--dry-run] [--rm] -> merge-back handler

# handler validates, compares branches, confirms mutations, executes git plan
merge-back handler -> git compare/rebase/merge -> optional worktree remove + branch -D
```

## Preconditions

- The `kool` command is available in PATH
- Git is available in PATH

## Steps

1. Verify `kool` and `git` are available
2. Execute `kool git worktree merge-back` with configured Request fields
3. Capture stdout, stderr, and exit code

## Context

- Source must be a linked worktree; main repo cwd is rejected
- Default target is the main repository; `--to` accepts main repo or sibling worktree under the same main repo
- Ahead and diverged paths require TTY confirmation or `--confirm-from-stdin`; non-TTY without that flag errors before mutations
- Already-included without `--rm` is a no-op success; `--rm` removes worktree and deletes branch without merge prompt
- `--dry-run` prints planned `git -C` commands and never mutates

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

func addDetachedSiblingWorktree(t *testing.T, mainRepo, wtName string) string {
	t.Helper()
	wtPath := filepath.Join(filepath.Dir(mainRepo), wtName)
	runGit(t, mainRepo, "worktree", "add", "--detach", wtPath, "HEAD")
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

func fileTrackedInRepo(t *testing.T, repoDir, filename string) bool {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "--error-unmatch", filename)
	cmd.Dir = repoDir
	return cmd.Run() == nil
}

func combinedOutput(resp *Response) string {
	return resp.Stdout + resp.Stderr
}

func outputContains(resp *Response, substr string) bool {
	return strings.Contains(combinedOutput(resp), substr)
}

func outputContainsAll(resp *Response, substrs ...string) bool {
	out := combinedOutput(resp)
	for _, s := range substrs {
		if !strings.Contains(out, s) {
			return false
		}
	}
	return true
}

func mainHasCommitMessage(t *testing.T, mainRepo, msg string) bool {
	t.Helper()
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = mainRepo
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log in %s: %v\n%s", mainRepo, err, out)
	}
	return strings.Contains(string(out), msg)
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

func rebaseInProgress(t *testing.T, wtPath string) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(wtPath, ".git", "rebase-merge"))
	if err == nil {
		return true
	}
	_, err = os.Stat(filepath.Join(wtPath, ".git", "rebase-apply"))
	return err == nil
}
```