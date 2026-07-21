# Scenario

**Feature**: trailing slash stripped from path in URI

```
# filepath.Clean removes trailing slash
validateGitRepoPath(path/) -> normalized without slash
```

## Steps
1. Create git repo.
2. Pass path with trailing `/`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markUriTree()
	markRootTree()
	repoDir := filepath.Join(req.WorkingDir, "repo")
	if err := osMkdir(repoDir); err != nil {
		return err
	}
	if err := initGitRepo(t, repoDir); err != nil {
		return err
	}
	req.RepoPath = repoDir + string(filepath.Separator)
	return nil
}
```