# Scenario

**Feature**: merge-back from detached HEAD worktree

```
# linked worktree checked out at commit ahead of main, not on a branch
worktree (detached HEAD) -> merge-back handler -> compare worktree commit vs main HEAD
```

## Steps

1. Create main repo and linked worktree on branch `feature`
2. Commit on feature without merging to main
3. Detach HEAD in the worktree (simulates `(no branch)` checkout)

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-detached", "feature")

	writeFile(t, filepath.Join(wtPath, "detached-ahead.txt"), "detached ahead work\n")
	runGit(t, wtPath, "add", ".")
	runGit(t, wtPath, "commit", "-m", "detached ahead commit")
	runGit(t, wtPath, "checkout", "--detach")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.TargetPath = mainRepo
	req.BranchName = "feature"
	req.Cwd = wtPath
	return nil
}
```