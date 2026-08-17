# Scenario

**Feature**: reclaim --all --dry-run leaves all worktrees intact

```
# dry-run suppresses all removals
user -> kool git worktree reclaim --all --dry-run -> would-reclaim only
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if !req.All || !req.DryRun {
		t.Fatal("expected reclaim-all dry-run setup from ancestors")
	}
	return nil
}
```