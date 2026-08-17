# Scenario

**Feature**: source branch is ahead of main HEAD

```
# feature has commits not reachable from main HEAD
merge-back handler -> compare branches -> ahead -> confirmation required
```

## Steps

1. Create main repo and linked worktree on branch `feature`
2. Commit on feature without merging to main

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-ahead", "feature")

	writeFile(t, filepath.Join(wtPath, "ahead.txt"), "ahead work\n")
	runGit(t, wtPath, "add", ".")
	runGit(t, wtPath, "commit", "-m", "feature ahead")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.TargetPath = mainRepo
	req.BranchName = "feature"
	req.Cwd = wtPath
	return nil
}
```