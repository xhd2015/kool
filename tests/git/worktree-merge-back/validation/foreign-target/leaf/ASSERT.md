## Expected

- Exit code non-zero
- Output mentions target does not share the same main repository

## Side Effects

- Source worktree still exists
- Foreign worktree still exists

## Errors

- Command rejects foreign target worktree

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
	if !strings.Contains(out, "same main") && !strings.Contains(out, "main repository") {
		t.Fatalf("expected error about foreign main repository, got:\n%s", out)
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected source worktree to remain: %s", req.WorktreePath)
	}
	if !pathExists(t, req.ForeignWT) {
		t.Fatalf("expected foreign worktree to remain: %s", req.ForeignWT)
	}
}
```