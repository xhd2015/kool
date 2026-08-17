# Scenario

**Feature**: reclaim --all removes dead worktree registration and deletes branch

```
# dead worktree is always reclaimable
reclaim handler -> git worktree remove --force + branch -D -> reclaimed: <path> (dead)
```

## Steps

1. Run reclaim --all without dry-run against the dead linked worktree

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorktreePath == "" || !req.All {
		t.Fatal("expected dead-worktree reclaim-all setup from ancestors")
	}
	req.DryRun = false
	return nil
}
```