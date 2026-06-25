## Expected

- Exit code 0
- Stdout contains `dry-run:` and `would reclaim` with the worktree path

## Side Effects

- Worktree directory still exists
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
	if !strings.Contains(out, req.WorktreePath) {
		t.Fatalf("expected worktree path in output, got:\n%s", out)
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree directory to remain: %s", req.WorktreePath)
	}
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain", req.BranchName)
	}
}
```