# Scenario

**Feature**: ahead dry-run uses branch name in merge command

```
# attached worktree on branch feature
merge-back handler -> build plan -> merge --ff-only feature (not commit hash)
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	req.Remove = false
	req.ConfirmFromStdin = false
	if req.WorktreePath == "" {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	if req.BranchName == "" {
		t.Fatal("expected branch name from ancestor setup")
	}
	return nil
}
```