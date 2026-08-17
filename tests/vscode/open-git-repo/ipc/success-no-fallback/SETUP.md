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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repoDir := initValidGitRepo(t, req.WorkingDir, "ipc-repo")
	req.RepoPath = repoDir
	return nil
}
```