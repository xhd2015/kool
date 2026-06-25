# Scenario

**Feature**: reclaim --all removes all reclaimable worktrees

```
# all candidates pass checks
user -> kool git worktree reclaim --all -> all reclaimed
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !req.All || req.MainRepo == "" {
		t.Fatal("expected reclaim-all all-reclaimable setup from ancestors")
	}
	return nil
}
```