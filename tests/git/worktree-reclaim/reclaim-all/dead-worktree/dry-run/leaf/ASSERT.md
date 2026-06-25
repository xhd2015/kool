## Expected

- Exit code 0
- Stdout contains `dry-run: would reclaim` with the worktree path and `(dead)`
- Stdout does not contain `reclaimed:` or `skipped:` with `uncommitted changes`

## Side Effects

- Worktree directory still does not exist
- Dead worktree entry still appears in `git worktree list`
- Branch `feature` still exists

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	out := combinedOutput(resp)
	if !strings.Contains(out, "dry-run:") || !strings.Contains(out, "would reclaim") {
		t.Fatalf("expected dry-run would-reclaim message, got:\n%s", out)
	}
	if !strings.Contains(out, req.WorktreePath) || !strings.Contains(out, "(dead)") {
		t.Fatalf("expected dead worktree path in output, got:\n%s", out)
	}
	if strings.Contains(out, "reclaimed:") {
		t.Fatalf("expected no reclaimed worktrees in dry-run, got:\n%s", out)
	}
	if strings.Contains(out, "skipped:") && strings.Contains(out, "uncommitted changes") {
		t.Fatalf("dead worktree must not be skipped as uncommitted changes, got:\n%s", out)
	}
	if pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree directory to remain absent: %s", req.WorktreePath)
	}
	if !worktreeListed(t, req.MainRepo, req.WorktreePath) {
		t.Fatalf("expected dead worktree still registered in git worktree list: %s", req.WorktreePath)
	}
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain", req.BranchName)
	}
}
```