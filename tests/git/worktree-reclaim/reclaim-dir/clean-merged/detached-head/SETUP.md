# Scenario

**Feature**: detached HEAD worktree with included commit is reclaimable

```
# worktree checked out at merged commit in detached HEAD state
reclaim handler -> compare detached HEAD commit against main HEAD -> reclaimable
```

## Context

- Detached HEAD worktrees compare the checkout commit directly, not a branch name

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected merged worktree from ancestor setup")
	}
	return nil
}
```