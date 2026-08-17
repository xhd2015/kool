# Scenario

**Feature**: dry-run leaves worktree and branch intact

```
# --dry-run reports would-reclaim
user -> kool git worktree reclaim <dir> --dry-run -> dry-run output only
```

## Steps

1. Enable dry-run for the merged feature worktree

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorktreePath == "" {
		t.Fatal("expected WorktreePath from ancestor setup")
	}
	req.DryRun = true
	return nil
}
```