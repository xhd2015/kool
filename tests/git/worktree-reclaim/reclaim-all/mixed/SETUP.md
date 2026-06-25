# Scenario

**Feature**: reclaim --all with mixed reclaimable and non-reclaimable worktrees

```
# one clean merged, one dirty linked worktree
reclaim handler -> reclaim one, skip one -> exit 0
```

## Steps

1. Create main repo with two linked worktrees
2. Merge `feature-a` into main; leave `feature-b` dirty

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)

	wtA := addLinkedWorktree(t, mainRepo, "wt-a", "feature-a")
	writeFile(t, filepath.Join(wtA, "a.txt"), "a work\n")
	runGit(t, wtA, "add", ".")
	runGit(t, wtA, "commit", "-m", "a work")
	mergeBranch(t, mainRepo, "feature-a")

	wtB := addLinkedWorktree(t, mainRepo, "wt-b", "feature-b")
	writeFile(t, filepath.Join(wtB, "b-dirty.txt"), "dirty\n")

	req.MainRepo = mainRepo
	req.WorktreePath = wtA
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```