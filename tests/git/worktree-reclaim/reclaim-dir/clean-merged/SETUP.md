# Scenario

**Feature**: clean worktree whose branch was merged into main is reclaimable

```
# feature branch merged into main; linked worktree still at included HEAD
reclaim handler -> git status (clean) + compare HEAD -> reclaimable
```

## Steps

1. Create main repo with initial commit
2. Add linked worktree on branch `feature`
3. Commit work on feature branch
4. Merge feature into main on main repo

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-feature", "feature")

	writeFile(t, filepath.Join(wtPath, "feature-work.txt"), "work\n")
	runGit(t, wtPath, "add", ".")
	runGit(t, wtPath, "commit", "-m", "feature work")

	mergeBranch(t, mainRepo, "feature")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.BranchName = "feature"
	req.Path = wtPath
	req.Cwd = mainRepo
	return nil
}
```