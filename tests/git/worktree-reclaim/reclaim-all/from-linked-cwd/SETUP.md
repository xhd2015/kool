# Scenario

**Feature**: reclaim --all from inside a linked worktree cwd

```
# cwd is linked worktree; handler resolves main repo automatically
user (cwd=linked wt) -> kool git worktree reclaim --all -> main repo resolved
```

## Steps

1. Create main repo with one merged linked worktree
2. Set Cwd to the linked worktree path

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-from-cwd", "feature")

	writeFile(t, filepath.Join(wtPath, "feature-work.txt"), "work\n")
	runGit(t, wtPath, "add", ".")
	runGit(t, wtPath, "commit", "-m", "feature work")
	mergeBranch(t, mainRepo, "feature")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.Cwd = wtPath
	req.DryRun = false
	return nil
}
```