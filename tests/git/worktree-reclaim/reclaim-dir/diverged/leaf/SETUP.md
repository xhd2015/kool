# Scenario

**Feature**: reclaim rejects diverged worktree

```
# branches diverged from main HEAD
user -> kool git worktree reclaim <diverged-wt> -> error
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatal("expected diverged worktree from ancestor setup")
	}
	return nil
}
```