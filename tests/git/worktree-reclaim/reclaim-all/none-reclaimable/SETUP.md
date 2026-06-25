# Scenario

**Feature**: reclaim --all when no worktrees are reclaimable

```
# all linked worktrees fail preconditions
reclaim handler -> skip all -> exit 0
```

## Steps

1. Create main repo with dirty and unmerged-ahead worktrees

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)

	wtDirty := addLinkedWorktree(t, mainRepo, "wt-dirty", "feature-dirty")
	writeFile(t, filepath.Join(wtDirty, "dirty.txt"), "dirty\n")

	wtAhead := addLinkedWorktree(t, mainRepo, "wt-ahead", "feature-ahead")
	writeFile(t, filepath.Join(wtAhead, "ahead.txt"), "ahead\n")
	runGit(t, wtAhead, "add", ".")
	runGit(t, wtAhead, "commit", "-m", "ahead work")

	req.MainRepo = mainRepo
	req.Cwd = mainRepo
	req.DryRun = false
	return nil
}
```