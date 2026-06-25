# Scenario

**Feature**: reclaim --all --dry-run leaves all worktrees intact

```
# dry-run suppresses all removals
user -> kool git worktree reclaim --all --dry-run -> would-reclaim only
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !req.All || !req.DryRun {
		t.Fatal("expected reclaim-all dry-run setup from ancestors")
	}
	return nil
}
```