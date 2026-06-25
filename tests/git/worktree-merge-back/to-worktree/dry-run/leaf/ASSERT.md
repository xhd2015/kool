## Expected

- Exit code 0
- Stdout lists planned `git -C` commands targeting sibling path
- No git mutations

## Side Effects

- Sibling worktree does not contain feature file
- Source worktree still exists

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
	if !strings.Contains(out, "git -C") || !strings.Contains(out, "merge") {
		t.Fatalf("expected dry-run output to list git -C merge command, got:\n%s", out)
	}
	if fileTrackedInRepo(t, req.SiblingPath, "sibling-ahead.txt") {
		t.Fatalf("expected sibling worktree not to contain sibling-ahead.txt after dry-run")
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected source worktree to remain: %s", req.WorktreePath)
	}
}
```