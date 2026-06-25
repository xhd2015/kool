# Scenario

**Feature**: merge-back rejects dirty linked worktree

```
# uncommitted changes block merge-back
merge-back handler -> git status (dirty) -> uncommitted changes error
```

## Steps

1. Create main repo and linked worktree on branch `feature`
2. Write an uncommitted file in the worktree

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
	req.Cwd = wtPath
	return nil
}
```