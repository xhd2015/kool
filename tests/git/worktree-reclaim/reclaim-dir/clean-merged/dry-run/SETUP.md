# Scenario

**Feature**: dry-run on reclaimable worktree reports without removing

```
# dry-run flag suppresses removal
reclaim handler -> dry-run: would reclaim <path> (no git worktree remove)
```

## Context

- Same merged clean worktree as sibling leaf; only dry-run differs

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