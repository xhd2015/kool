# Scenario

**Feature**: reclaim --all --dry-run reports dead worktree without removing registration

```
# dead worktree still registered; dry-run only
user -> kool git worktree reclaim --all --dry-run -> dry-run: would reclaim <path> (dead)
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorktreePath == "" || !req.All {
		t.Fatal("expected dead-worktree reclaim-all setup from ancestors")
	}
	req.DryRun = true
	return nil
}
```