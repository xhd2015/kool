# Scenario

**Feature**: merge-back targets a sibling worktree via --to

```
# source linked wt ahead of sibling HEAD
user -> merge-back --to <sibling-wt> -> merge into sibling checkout
```

## Steps

1. Create main repo with detached sibling worktree at main HEAD
2. Add source linked worktree on branch `feature`
3. Commit on feature so branch is ahead of sibling HEAD

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	siblingPath := addDetachedSiblingWorktree(t, mainRepo, "wt-sibling")
	wtPath := addLinkedWorktree(t, mainRepo, "wt-source", "feature")

	writeFile(t, filepath.Join(wtPath, "sibling-ahead.txt"), "ahead for sibling\n")
	runGit(t, wtPath, "add", ".")
	runGit(t, wtPath, "commit", "-m", "feature ahead for sibling target")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.SiblingPath = siblingPath
	req.TargetPath = siblingPath
	req.BranchName = "feature"
	req.Cwd = wtPath
	req.To = siblingPath
	return nil
}
```