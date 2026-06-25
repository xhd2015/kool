## Expected

- Exit code non-zero
- Output mentions non-terminal stdin or confirmation required

## Side Effects

- Worktree directory still exists
- Branch still exists
- Diverged state preserved

## Errors

- Non-interactive stdin cannot confirm rebase+merge

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
	if !strings.Contains(out, "terminal") && !strings.Contains(out, "confirm") {
		t.Fatalf("expected error about non-TTY or confirmation, got:\n%s", out)
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree to remain: %s", req.WorktreePath)
	}
	if rebaseInProgress(t, req.WorktreePath) {
		t.Fatalf("expected no rebase in progress after non-tty rejection")
	}
}
```