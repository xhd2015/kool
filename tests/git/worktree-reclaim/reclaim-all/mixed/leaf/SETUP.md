# Scenario

**Feature**: mixed reclaim-all reclaims eligible worktree and skips dirty one

```
# --all processes each linked worktree independently
user -> kool git worktree reclaim --all -> one reclaimed, one skipped
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !req.All || req.WorktreePath == "" {
		t.Fatal("expected reclaim-all mixed setup from ancestors")
	}
	return nil
}
```