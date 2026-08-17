# Scenario

**Feature**: reclaim --all from linked worktree cwd reclaims eligible worktrees

```
# main repo resolved from linked worktree .git gitdir
user (cwd=wt) -> kool git worktree reclaim --all -> reclaimed
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.Cwd != req.WorktreePath {
		t.Fatalf("expected Cwd inside linked worktree, cwd=%q wt=%q", req.Cwd, req.WorktreePath)
	}
	return nil
}
```