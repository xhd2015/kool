## Expected

- Exit code 0
- Rebase+merge succeeds; worktree removed; branch deleted

## Side Effects

- Main repo contains both feature and main changes
- Worktree directory no longer exists
- Branch `feature` deleted

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if !fileTrackedInRepo(t, req.MainRepo, "feature.txt") {
		t.Fatalf("expected main repo to contain feature.txt after rebase+merge")
	}
	if pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree removed: %s", req.WorktreePath)
	}
	if branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q deleted", req.BranchName)
	}
}
```