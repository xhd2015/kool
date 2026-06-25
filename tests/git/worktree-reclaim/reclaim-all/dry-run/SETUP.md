# Scenario

**Feature**: reclaim --all --dry-run reports without removing

```
# all linked worktrees are reclaimable but dry-run is set
reclaim handler -> dry-run messages only
```

## Steps

1. Create main repo with two merged linked worktrees

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
	_ = setupMergedWorktree(t, mainRepo, "wt-a", "feature-a")
	_ = setupMergedWorktree(t, mainRepo, "wt-b", "feature-b")

	req.MainRepo = mainRepo
	req.Cwd = mainRepo
	req.DryRun = true
	return nil
}
```