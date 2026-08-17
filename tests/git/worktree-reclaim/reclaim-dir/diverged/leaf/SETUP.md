# Scenario

**Feature**: reclaim rejects diverged worktree

```
# branches diverged from main HEAD
user -> kool git worktree reclaim <diverged-wt> -> error
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatal("expected diverged worktree from ancestor setup")
	}
	return nil
}
```