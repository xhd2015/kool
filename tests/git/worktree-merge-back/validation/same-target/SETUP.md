# Scenario

**Feature**: merge-back rejects when source and target resolve to the same worktree

```
# --to points at the same linked worktree as source
user -> merge-back --to <same-wt> -> source and target are the same worktree
```

## Steps

1. Create main repo and linked worktree on branch `feature`
2. Set `--to` to the source worktree path

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-same", "feature")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.BranchName = "feature"
	req.Cwd = wtPath
	req.To = wtPath
	return nil
}
```