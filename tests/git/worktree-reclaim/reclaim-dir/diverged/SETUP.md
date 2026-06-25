# Scenario

**Feature**: worktree branch diverged from main HEAD is not reclaimable

```
# both main and feature have unique commits
reclaim handler -> compare branches -> diverged -> error
```

## Steps

1. Create main repo and linked worktree on branch `feature`
2. Commit on feature, then commit on main creating divergence

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-diverged", "feature")

	writeFile(t, filepath.Join(wtPath, "feature.txt"), "feature change\n")
	runGit(t, wtPath, "add", ".")
	runGit(t, wtPath, "commit", "-m", "feature change")

	writeFile(t, filepath.Join(mainRepo, "main.txt"), "main change\n")
	runGit(t, mainRepo, "add", ".")
	runGit(t, mainRepo, "commit", "-m", "main change")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.BranchName = "feature"
	req.Path = wtPath
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```