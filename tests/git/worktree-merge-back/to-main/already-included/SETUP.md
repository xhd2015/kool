# Scenario

**Feature**: source branch already included in main HEAD

```
# feature merged into main; worktree HEAD is ancestor of main HEAD
merge-back handler -> compare branches -> already included
```

## Steps

1. Create main repo and linked worktree on branch `feature`
2. Commit on feature, merge feature into main

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-included", "feature")

	writeFile(t, filepath.Join(wtPath, "feature-work.txt"), "work\n")
	runGit(t, wtPath, "add", ".")
	runGit(t, wtPath, "commit", "-m", "feature work")

	mergeBranch(t, mainRepo, "feature")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.TargetPath = mainRepo
	req.BranchName = "feature"
	req.Cwd = wtPath
	return nil
}
```