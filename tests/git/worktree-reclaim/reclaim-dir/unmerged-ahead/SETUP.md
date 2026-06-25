# Scenario

**Feature**: worktree branch ahead of main HEAD is not reclaimable

```
# feature has commits not reachable from main HEAD
reclaim handler -> compare branches -> not included -> error
```

## Steps

1. Create main repo and linked worktree on branch `feature`
2. Commit on feature without merging to main

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-ahead", "feature")

	writeFile(t, filepath.Join(wtPath, "ahead.txt"), "ahead work\n")
	runGit(t, wtPath, "add", ".")
	runGit(t, wtPath, "commit", "-m", "feature ahead")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.BranchName = "feature"
	req.Path = wtPath
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```