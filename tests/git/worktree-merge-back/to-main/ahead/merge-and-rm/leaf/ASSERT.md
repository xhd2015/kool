## Expected

- Exit code 0
- Prompt lists remove and branch delete commands
- Target merged; worktree removed; branch deleted

## Side Effects

- Main repo contains `ahead.txt`
- Worktree directory no longer exists
- Branch `feature` deleted

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
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "worktree remove") && !strings.Contains(lower, "worktree remove") {
		if !strings.Contains(lower, "remove") {
			t.Fatalf("expected prompt to list worktree remove, got:\n%s", out)
		}
	}
	if !strings.Contains(lower, "branch -d") && !strings.Contains(lower, "branch -d") {
		if !strings.Contains(lower, "branch") {
			t.Fatalf("expected prompt to list branch delete, got:\n%s", out)
		}
	}
	if !fileTrackedInRepo(t, req.MainRepo, "ahead.txt") {
		t.Fatalf("expected main repo to contain ahead.txt after merge")
	}
	if pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree removed: %s", req.WorktreePath)
	}
	if branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q deleted", req.BranchName)
	}
}
```