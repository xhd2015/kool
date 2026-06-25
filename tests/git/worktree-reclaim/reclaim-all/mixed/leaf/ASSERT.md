## Expected

- Exit code 0
- Stdout contains `reclaimed:` for the merged worktree
- Stdout contains `skipped:` for the dirty worktree

## Side Effects

- Merged worktree `wt-a` is removed
- Dirty worktree `wt-b` still exists

## Exit Code

- 0

```go
import (
	"path/filepath"
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
	wtB := filepath.Join(filepath.Dir(req.MainRepo), "wt-b")

	if !strings.Contains(out, "reclaimed:") || !strings.Contains(out, req.WorktreePath) {
		t.Fatalf("expected reclaimed message for %s, got:\n%s", req.WorktreePath, out)
	}
	if !strings.Contains(out, "skipped:") || !strings.Contains(out, wtB) {
		t.Fatalf("expected skipped message for %s, got:\n%s", wtB, out)
	}
	if pathExists(t, req.WorktreePath) {
		t.Fatalf("expected merged worktree removed: %s", req.WorktreePath)
	}
	if !pathExists(t, wtB) {
		t.Fatalf("expected dirty worktree to remain: %s", wtB)
	}
}
```