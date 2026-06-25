# Scenario

**Feature**: reclaim clean merged worktree removes directory and deletes branch

```
# reclaimable candidate passes checks
reclaim handler -> git worktree remove + branch -D -> reclaimed: <path>
```

## Steps

1. Target the merged feature worktree without dry-run

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorktreePath == "" {
		t.Fatal("expected WorktreePath from ancestor setup")
	}
	req.DryRun = false
	return nil
}
```