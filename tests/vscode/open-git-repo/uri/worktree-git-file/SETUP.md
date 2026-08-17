# Scenario

**Feature**: worktree with `.git` file is accepted for URI building

```
# .git file (not directory) is valid git metadata
validateGitRepoPath(worktree) -> buildGitOpenRepoURI
```

## Steps
1. Create main git repo and worktree with `.git` file.
2. Build URI from worktree path.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := filepath.Join(req.WorkingDir, "main")
	worktree := filepath.Join(req.WorkingDir, "worktree")
	if err := osMkdir(mainRepo); err != nil {
		return err
	}
	if err := initGitRepo(t, mainRepo); err != nil {
		return err
	}
	if err := osMkdir(worktree); err != nil {
		return err
	}
	gitdir := filepath.Join(mainRepo, ".git")
	if err := os.WriteFile(filepath.Join(worktree, ".git"), []byte("gitdir: "+gitdir+"\n"), 0644); err != nil {
		return err
	}
	req.RepoPath = worktree
	return nil
}
```