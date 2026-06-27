# Scenario

**Feature**: `open` command invoked with vscode:// URI on darwin

```
# mocked exec captures open <uri>
OpenGitRepo(validRepo) -> exec("open", uri)
```

## Steps
1. Create valid git repo.
2. Call `OpenGitRepo` with mocked exec on darwin.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Phase = "exec"
	req.GoOS = "darwin"
	repoDir := filepath.Join(req.WorkingDir, "repo")
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