# Scenario

**Feature**: diverged branches with conflicting changes cause rebase failure

```
# same file edited differently on main and feature
merge-back handler -> rebase -> CONFLICT -> abort rebase, error
```

## Steps

1. Create diverged repos where README.md conflicts on rebase

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	wtPath := addLinkedWorktree(t, mainRepo, "wt-conflict", "feature")

	writeFile(t, filepath.Join(wtPath, "README.md"), "# feature change\n")
	runGit(t, wtPath, "add", "README.md")
	runGit(t, wtPath, "commit", "-m", "feature change to README")

	writeFile(t, filepath.Join(mainRepo, "README.md"), "# main change\n")
	runGit(t, mainRepo, "add", "README.md")
	runGit(t, mainRepo, "commit", "-m", "main change to README")

	req.MainRepo = mainRepo
	req.WorktreePath = wtPath
	req.TargetPath = mainRepo
	req.BranchName = "feature"
	req.Cwd = wtPath
	return nil
}
```