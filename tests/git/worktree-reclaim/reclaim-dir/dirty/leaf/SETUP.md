# Scenario

**Feature**: reclaim rejects dirty worktree

```
# single-path reclaim on dirty worktree fails
user -> kool git worktree reclaim <dirty-wt> -> error
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatal("expected dirty worktree from ancestor setup")
	}
	return nil
}
```