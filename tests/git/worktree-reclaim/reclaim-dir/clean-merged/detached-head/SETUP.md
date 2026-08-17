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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatalf("expected merged worktree from ancestor setup")
	}
	return nil
}
```