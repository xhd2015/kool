# Scenario

**Feature**: reclaim --all removes every reclaimable linked worktree

```
# two clean merged worktrees
reclaim handler -> remove all linked worktrees
```

## Steps

1. Create main repo with two linked worktrees, merge both into main

```go
import (
	"path/filepath"
	"testing"
)

func setupMergedWorktree(t *testing.T, mainRepo, wtName, branch string) string {
	t.Helper()
	wtPath := addLinkedWorktree(t, mainRepo, wtName, branch)
	writeFile(t, filepath.Join(wtPath, branch+".txt"), "work\n")
	runGit(t, wtPath, "add", ".")
	runGit(t, wtPath, "commit", "-m", branch+" work")
	mergeBranch(t, mainRepo, branch)
	return wtPath
}

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtA := setupMergedWorktree(t, mainRepo, "wt-a", "feature-a")
	_ = setupMergedWorktree(t, mainRepo, "wt-b", "feature-b")

	req.MainRepo = mainRepo
	req.WorktreePath = wtA
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```