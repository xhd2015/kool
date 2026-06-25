## Expected

- Exit code 0
- Target fast-forward merged; source worktree remains

## Side Effects

- Main repo contains `ahead.txt` from feature branch
- Worktree directory still exists
- Branch `feature` still exists

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
	if !strings.Contains(strings.ToLower(out), "ahead") && !strings.Contains(out, "merge") {
		t.Fatalf("expected output to mention ahead merge, got:\n%s", out)
	}
	if !fileTrackedInRepo(t, req.MainRepo, "ahead.txt") {
		t.Fatalf("expected main repo to contain ahead.txt after merge")
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree to remain: %s", req.WorktreePath)
	}
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain without --rm", req.BranchName)
	}
}
```