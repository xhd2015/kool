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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repoDir := initValidGitRepo(t, req.WorkingDir, "fallback-repo")
	req.RepoPath = repoDir
	req.IPCAlwaysFail = true
	return nil
}
```