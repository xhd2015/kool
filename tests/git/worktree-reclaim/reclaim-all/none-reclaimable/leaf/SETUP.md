# Scenario

**Feature**: reclaim --all skips all non-reclaimable worktrees

```
# no removals attempted successfully
user -> kool git worktree reclaim --all -> all skipped, exit 0
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !req.All || req.MainRepo == "" {
		t.Fatal("expected reclaim-all none-reclaimable setup from ancestors")
	}
	return nil
}
```