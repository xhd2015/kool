# Scenario

**Feature**: reclaim --all from linked worktree cwd reclaims eligible worktrees

```
# main repo resolved from linked worktree .git gitdir
user (cwd=wt) -> kool git worktree reclaim --all -> reclaimed
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.Cwd != req.WorktreePath {
		t.Fatalf("expected Cwd inside linked worktree, cwd=%q wt=%q", req.Cwd, req.WorktreePath)
	}
	return nil
}
```