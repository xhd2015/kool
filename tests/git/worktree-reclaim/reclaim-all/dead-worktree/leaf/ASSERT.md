## Expected

- Exit code 0
- Stdout contains `reclaimed:` with the worktree path and `(dead)`
- Stdout does not contain `skipped:` with `uncommitted changes`

## Side Effects

- Worktree directory still does not exist
- Dead worktree entry removed from `git worktree list`
- Branch `feature` is deleted from main repo

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
	if !strings.Contains(out, "reclaimed:") || !strings.Contains(out, req.WorktreePath) || !strings.Contains(out, "(dead)") {
		t.Fatalf("expected reclaimed dead worktree path in output, got:\n%s", out)
	}
	if strings.Contains(out, "skipped:") && strings.Contains(out, "uncommitted changes") {
		t.Fatalf("dead worktree must not be skipped as uncommitted changes, got:\n%s", out)
	}
	if pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree directory to remain absent: %s", req.WorktreePath)
	}
	if worktreeListed(t, req.MainRepo, req.WorktreePath) {
		t.Fatalf("expected dead worktree removed from git worktree list: %s", req.WorktreePath)
	}
	if branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q deleted", req.BranchName)
	}
}
```