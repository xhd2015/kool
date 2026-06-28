# Scenario

**Feature**: IPC failure triggers stderr hint and git-open URI fallback

```
# no IPC server; retries exhausted
OpenGitRepo -> IPC (fail) -> stderr hint -> OS opener(vscode:// URI)
```

## Steps
1. Create valid git repo.
2. Skip IPC server (`IPCAlwaysFail`).
3. Call `OpenGitRepo` with exec hook on darwin.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	repoDir := initValidGitRepo(t, req.WorkingDir, "fallback-repo")
	req.RepoPath = repoDir
	req.IPCAlwaysFail = true
	return nil
}
```