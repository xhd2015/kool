# Scenario

**Feature**: reclaim --all skips all non-reclaimable worktrees

```
# no removals attempted successfully
user -> kool git worktree reclaim --all -> all skipped, exit 0
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if !req.All || req.MainRepo == "" {
		t.Fatal("expected reclaim-all none-reclaimable setup from ancestors")
	}
	return nil
}
```