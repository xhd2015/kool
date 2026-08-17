# Scenario

**Feature**: single-path reclaim removes dead worktree registration

```
# dead worktree path resolves via main repo worktree list
reclaim handler -> git worktree remove --force + branch -D -> reclaimed: <path> (dead)
```

## Steps

1. Target the dead linked worktree path without dry-run

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorktreePath == "" {
		t.Fatal("expected WorktreePath from dead-worktree ancestor setup")
	}
	req.DryRun = false
	return nil
}
```