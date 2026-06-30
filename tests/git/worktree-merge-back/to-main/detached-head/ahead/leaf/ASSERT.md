## Expected

- Exit code 0
- Detached HEAD commit is merged into main; not reported as already-included

## Side Effects

- Main repo contains `detached-ahead.txt` from the detached worktree commit
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
	if strings.Contains(strings.ToLower(out), "already included") {
		t.Fatalf("detached HEAD ahead of main must not be reported as already included, got:\n%s", out)
	}
	if !fileTrackedInRepo(t, req.MainRepo, "detached-ahead.txt") {
		t.Fatalf("expected main repo to contain detached-ahead.txt after merge-back; output:\n%s", out)
	}
	if !mainHasCommitMessage(t, req.MainRepo, "detached ahead commit") {
		t.Fatalf("expected main repo to include detached ahead commit; output:\n%s", out)
	}
	if !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree to remain: %s", req.WorktreePath)
	}
	if !branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q to remain without --rm", req.BranchName)
	}
}
```