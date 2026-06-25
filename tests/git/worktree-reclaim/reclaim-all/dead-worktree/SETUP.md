# Scenario

**Feature**: reclaim --all reclaims dead linked worktrees whose directory was deleted

```
# linked worktree path remains in git worktree list but dir is gone
user -> rm -rf <worktree> -> kool git worktree reclaim --all -> reclaimed: <path> (dead)
```

## Steps

1. Create main repo with linked worktree on branch `feature`
2. Delete the worktree directory from the filesystem (git registration remains)

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-dead", "feature")
	deleteWorktreeDir(t, wtPath)

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.BranchName = "feature"
	req.Cwd = mainRepo
	return nil
}
```