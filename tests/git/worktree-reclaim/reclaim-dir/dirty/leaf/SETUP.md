# Scenario

**Feature**: reclaim rejects dirty worktree

```
# single-path reclaim on dirty worktree fails
user -> kool git worktree reclaim <dirty-wt> -> error
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatal("expected dirty worktree from ancestor setup")
	}
	return nil
}
```