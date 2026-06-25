## Expected

- Exit code non-zero
- Output mentions not a linked worktree

## Side Effects

- Main repository remains unchanged

## Errors

- Command rejects main repo cwd as source

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
	if !strings.Contains(out, "linked worktree") && !strings.Contains(out, "not a linked") {
		t.Fatalf("expected error about not a linked worktree, got:\n%s", out)
	}
}
```