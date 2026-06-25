## Expected

- Exit code 0
- Stdout contains `reclaimed:` for both worktrees

## Side Effects

- Both linked worktree directories are removed
- Branches `feature-a` and `feature-b` are deleted

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

	if strings.Count(out, "reclaimed:") < 2 {
		t.Fatalf("expected two reclaimed messages, got:\n%s", out)
	}
	if !strings.Contains(out, wtA) || !strings.Contains(out, wtB) {
		t.Fatalf("expected both worktree paths in output, got:\n%s", out)
	}
	if pathExists(t, wtA) || pathExists(t, wtB) {
		t.Fatalf("expected both worktrees removed")
	}
	if branchExists(t, req.MainRepo, "feature-a") || branchExists(t, req.MainRepo, "feature-b") {
		t.Fatalf("expected branches deleted")
	}
}
```