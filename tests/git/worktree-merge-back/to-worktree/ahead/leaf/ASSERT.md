## Expected

- Exit code 0
- Sibling worktree HEAD includes feature commit
- Source worktree remains

## Side Effects

- `sibling-ahead.txt` tracked in sibling worktree checkout
- Source worktree directory still exists
- Branch `feature` still exists

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if !fileTrackedInRepo(t, req.SiblingPath, "sibling-ahead.txt") {
		t.Fatalf("expected sibling worktree to contain sibling-ahead.txt after merge")
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected source worktree to remain: %s", req.WorktreePath)
	}
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain", req.BranchName)
	}
}
```