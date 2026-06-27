# Scenario

**Feature**: relative path resolved against cwd before URI encoding

```
# relative path joined to cwd
validateGitRepoPath(relative, cwd) -> absolute in URI
```

## Steps
1. Create git repo at `repo` under working dir.
2. Pass relative path `repo` to URI builder.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	repoDir := filepath.Join(req.WorkingDir, "repo")
	if err := osMkdir(repoDir); err != nil {
		return err
	}
	if err := initGitRepo(t, repoDir); err != nil {
		return err
	}
	req.RepoPath = "repo"
	return nil
}
```