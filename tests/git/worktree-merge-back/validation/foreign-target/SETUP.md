# Scenario

**Feature**: merge-back rejects --to worktree from a different main repository

```
# foreign target does not share main repo with source
user -> merge-back --to <foreign-wt> -> target does not share the same main repository
```

## Steps

1. Create source main repo with linked worktree
2. Create separate foreign main repo with its own linked worktree
3. Run merge-back from source worktree with `--to` foreign worktree

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-source", "feature")

	foreignMain := initMainRepo(t)
	foreignWT := addLinkedWorktree(t, foreignMain, "wt-foreign", "other")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.BranchName = "feature"
	req.ForeignMain = foreignMain
	req.ForeignWT = foreignWT
	req.Cwd = wtPath
	req.To = foreignWT
	return nil
}
```