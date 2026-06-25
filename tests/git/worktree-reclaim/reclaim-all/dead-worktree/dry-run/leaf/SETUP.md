# Scenario

**Feature**: dry-run leaves dead worktree registered in git

```
# --dry-run suppresses removal of dead worktree entry
reclaim handler -> dry-run: would reclaim <path> (dead) -> git worktree list unchanged
```

## Steps

1. Run reclaim --all --dry-run against the dead linked worktree

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !req.DryRun || req.WorktreePath == "" {
		t.Fatal("expected dead-worktree dry-run setup from ancestors")
	}
	return nil
}
```