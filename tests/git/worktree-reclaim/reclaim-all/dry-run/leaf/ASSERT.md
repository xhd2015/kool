## Expected

- Exit code 0
- Stdout contains `dry-run:` and `would reclaim` for each worktree
- Stdout does not contain `reclaimed:`

## Side Effects

- Both linked worktree directories still exist
- Branches still exist

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
	wtA := filepath.Join(filepath.Dir(req.MainRepo), "wt-a")
	wtB := filepath.Join(filepath.Dir(req.MainRepo), "wt-b")

	if strings.Contains(out, "reclaimed:") {
		t.Fatalf("expected no reclaimed worktrees in dry-run, got:\n%s", out)
	}
	if strings.Count(out, "would reclaim") < 2 {
		t.Fatalf("expected two would-reclaim messages, got:\n%s", out)
	}
	if !pathExists(t, wtA) || !pathExists(t, wtB) {
		t.Fatalf("expected both worktrees to remain")
	}
	if !branchExists(t, req.MainRepo, "feature-a") || !branchExists(t, req.MainRepo, "feature-b") {
		t.Fatalf("expected branches to remain")
	}
}
```