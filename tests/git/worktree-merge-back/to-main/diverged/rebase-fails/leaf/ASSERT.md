## Expected

- Exit code non-zero
- Rebase aborted; source worktree unchanged on feature branch

## Side Effects

- Worktree directory still exists
- Branch still exists
- Main repo does not contain rebased feature README content
- No rebase in progress after abort

## Errors

- Rebase conflict reported

## Exit Code

- 1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0\nstdout: %s\nstderr: %s", resp.Stdout, resp.Stderr)
	}
	out := strings.ToLower(combinedOutput(resp))
	if !strings.Contains(out, "conflict") && !strings.Contains(out, "rebase") {
		t.Fatalf("expected error about rebase conflict, got:\n%s", out)
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree to remain: %s", req.WorktreePath)
	}
	if rebaseInProgress(t, req.WorktreePath) {
		t.Fatalf("expected rebase aborted with no rebase state left")
	}
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain after failed rebase", req.BranchName)
	}
}
```