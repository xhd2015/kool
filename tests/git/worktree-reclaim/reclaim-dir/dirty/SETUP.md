# Scenario

**Feature**: dirty linked worktree is not reclaimable

```
# uncommitted changes fail cleanliness check
reclaim handler -> git status --porcelain (non-empty) -> error
```

## Steps

1. Create main repo and linked worktree on branch `feature`
2. Modify a file without committing

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-dirty", "feature")

	writeFile(t, filepath.Join(wtPath, "dirty.txt"), "uncommitted\n")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.BranchName = "feature"
	req.Path = wtPath
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```