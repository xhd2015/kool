# Scenario

**Feature**: reclaim --all removes all reclaimable worktrees

```
# all candidates pass checks
user -> kool git worktree reclaim --all -> all reclaimed
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if !req.All || req.MainRepo == "" {
		t.Fatal("expected reclaim-all all-reclaimable setup from ancestors")
	}
	return nil
}
```