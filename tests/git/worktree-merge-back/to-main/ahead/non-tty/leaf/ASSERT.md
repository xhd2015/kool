## Expected

- Exit code non-zero
- Output mentions non-terminal stdin or confirmation required

## Side Effects

- Worktree directory still exists
- Branch still exists
- Main repo does not contain feature ahead commit

## Errors

- Non-interactive stdin cannot confirm ahead merge

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
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain", req.BranchName)
	}
	if fileTrackedInRepo(t, req.MainRepo, "ahead.txt") {
		t.Fatalf("expected main repo not to contain ahead.txt before merge")
	}
}
```