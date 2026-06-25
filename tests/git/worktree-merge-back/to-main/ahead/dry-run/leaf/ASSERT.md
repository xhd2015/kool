## Expected

- Exit code 0
- Stdout lists planned `git -C` merge command
- No git mutations

## Side Effects

- Worktree directory still exists
- Branch still exists
- Main repo does not contain ahead commit

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
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree to remain: %s", req.WorktreePath)
	}
	if fileTrackedInRepo(t, req.MainRepo, "ahead.txt") {
		t.Fatalf("expected main repo not to contain ahead.txt after dry-run")
	}
}
```