## Expected

- Exit code non-zero
- Output mentions not a worktree, linked worktree, or main repo

## Errors

- Command rejects path that is not a linked worktree

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
	if !strings.Contains(out, "worktree") && !strings.Contains(out, "linked") {
		t.Fatalf("expected error about not being a linked worktree, got:\n%s", out)
	}
	if !pathExists(t, req.MainRepo) {
		t.Fatalf("expected main repo to remain: %s", req.MainRepo)
	}
}
```