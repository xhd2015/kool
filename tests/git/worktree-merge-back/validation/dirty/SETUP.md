# Scenario

**Feature**: merge-back --rm rejects a dirty linked worktree

```
# uncommitted changes block --rm (worktree would be deleted)
merge-back --rm -> git status (dirty) -> uncommitted changes error
```

## Steps

1. Create main repo and linked worktree on branch `feature`
2. Write an uncommitted file in the worktree
3. Set `--rm` — dirty is only an error when the worktree would be removed

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-dirty", "feature")
	writeFile(t, filepath.Join(wtPath, "dirty.txt"), "uncommitted\n")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.BranchName = "feature"
	req.Cwd = wtPath
	req.Remove = true
	return nil
}
```