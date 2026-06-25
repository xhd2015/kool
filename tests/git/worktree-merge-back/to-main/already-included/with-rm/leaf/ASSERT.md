## Expected

- Exit code 0
- Worktree removed and branch deleted without merge prompt

## Side Effects

- Worktree directory no longer exists
- Branch `feature` deleted from main repo
- Main repo still contains merged feature commits

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
	if pathExists(t, req.WorktreePath) {
		t.Fatalf("expected worktree directory removed: %s", req.WorktreePath)
	}
	if branchExists(t, req.MainRepo, req.BranchName) {
		t.Fatalf("expected branch %q deleted", req.BranchName)
	}
	if !mainHasCommitMessage(t, req.MainRepo, "feature work") {
		t.Fatalf("expected main repo to retain feature commit")
	}
}
```