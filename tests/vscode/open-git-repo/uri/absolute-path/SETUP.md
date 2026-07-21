# Scenario

**Feature**: absolute git repo path produces correct vscode:// URI

```
# absolute input normalized and encoded
validateGitRepoPath(absPath) -> buildGitOpenRepoURI
```

## Steps
1. Create git repo at absolute path under temp dir.
2. Build URI from absolute path.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markUriTree()
	markRootTree()
	repoDir := filepath.Join(req.WorkingDir, "abs-repo")
	if err := osMkdir(repoDir); err != nil {
		return err
	}
	if err := initGitRepo(t, repoDir); err != nil {
		return err
	}
	req.RepoPath = repoDir
	return nil
}
```