## Expected

- Exit code 0
- No merge or removal performed

## Side Effects

- Worktree directory still exists
- Branch `feature` still exists
- Main repo HEAD unchanged relative to pre-run state

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree to remain: %s", req.WorktreePath)
	}
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain", req.BranchName)
	}
	if !worktreeListed(t, req.MainRepo, req.WorktreePath) {
		t.Fatalf("expected worktree still listed: %s", req.WorktreePath)
	}
}
```