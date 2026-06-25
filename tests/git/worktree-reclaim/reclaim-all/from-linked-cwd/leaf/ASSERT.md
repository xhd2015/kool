## Expected

- Exit code 0
- Stdout contains `reclaimed:` and the worktree path

## Side Effects

- Linked worktree directory is removed (reclaiming self from its own cwd)
- Branch `feature` is deleted

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
	if !strings.Contains(out, "reclaimed:") || !strings.Contains(out, req.WorktreePath) {
		t.Fatalf("expected reclaimed worktree path, got:\n%s", out)
	}
	if pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree directory removed: %s", req.WorktreePath)
	}
	if branchExists(t, req.MainRepo, "feature") {
		t.Fatalf("expected branch feature deleted")
	}
}
```