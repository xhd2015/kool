## Expected

- Exit code 0
- Stdout contains `skipped:` for each linked worktree
- Stdout does not contain `reclaimed:`

## Side Effects

- All linked worktree directories still exist

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
	wtDirty := filepath.Join(filepath.Dir(req.MainRepo), "wt-dirty")
	wtAhead := filepath.Join(filepath.Dir(req.MainRepo), "wt-ahead")

	if strings.Contains(out, "reclaimed:") {
		t.Fatalf("expected no reclaimed worktrees, got:\n%s", out)
	}
	if !strings.Contains(out, "skipped:") {
		t.Fatalf("expected skipped messages, got:\n%s", out)
	}
	if !pathExists(t, wtDirty) || !pathExists(t, wtAhead) {
		t.Fatalf("expected all worktrees to remain")
	}
}
```