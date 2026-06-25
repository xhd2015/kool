## Expected

- Exit code non-zero
- Output mentions not included, ahead, or not merged

## Side Effects

- Worktree directory still exists

## Errors

- Command reports worktree HEAD is not included in main HEAD

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
	if !strings.Contains(out, "included") && !strings.Contains(out, "ahead") && !strings.Contains(out, "merged") {
		t.Fatalf("expected error about HEAD not included in main, got:\n%s", out)
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree directory to remain: %s", req.WorktreePath)
	}
}
```