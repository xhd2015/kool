# Scenario

**Feature**: reclaim single dead linked worktree by path

```
# path missing on disk but still in git worktree list
user -> rm -rf <worktree> -> kool git worktree reclaim <path> -> reclaimed: <path> (dead)
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
	req.Path = wtPath
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```