# Scenario

**Feature**: spaces in path are URL-encoded in URI

```
# path with spaces encoded as %20
buildGitOpenRepoURI("/tmp/my repo") -> path=%2Ftmp%2Fmy%20repo
```

## Steps
1. Create git repo in directory with spaces in name.
2. Build URI from that path.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	repoDir := filepath.Join(req.WorkingDir, "my repo")
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