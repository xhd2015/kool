## Expected

- Exit code non-zero
- Output mentions uncommitted changes

## Side Effects

- Worktree directory still exists
- Branch `feature` still exists

## Errors

- Command reports worktree has uncommitted changes

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
	if !strings.Contains(out, "uncommitted") {
		t.Fatalf("expected error about uncommitted changes, got:\n%s", out)
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree directory to remain: %s", req.WorktreePath)
	}
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain", req.BranchName)
	}
}
```