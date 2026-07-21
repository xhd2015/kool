# Scenario

**Feature**: IPC success opens git repo without OS fallback

```
# IPC responds ok; exec hook not invoked
OpenGitRepo -> IPC {"op":"git-open"} -> ok:true
```

## Steps
1. Create valid git repo.
2. Start mock IPC server.
3. Call `OpenGitRepo`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markIpcTree()
	markRootTree()
	repoDir := initValidGitRepo(t, req.WorkingDir, "ipc-repo")
	req.RepoPath = repoDir
	return nil
}
```