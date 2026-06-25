## Expected

- Exit code 0
- Operation aborted without git mutations

## Side Effects

- Worktree directory still exists
- Branch still exists
- Main repo does not contain feature ahead commit

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
		t.Fatalf("expected exit code 0 on decline, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree to remain: %s", req.WorktreePath)
	}
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain", req.BranchName)
	}
	if fileTrackedInRepo(t, req.MainRepo, "ahead.txt") {
		t.Fatalf("expected main repo not to contain ahead.txt after decline")
	}
}
```