# Scenario

**Feature**: missing `code` CLI blocks open-git-repo

```
# code not on PATH
EnsureCodeCLI -> error (mentions code / PATH)
```

## Steps
1. Create valid git repo.
2. Run CLI with empty PATH.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repoDir := initValidGitRepo(t, req.WorkingDir, "repo")
	req.RepoPath = repoDir
	req.CodeInPath = false
	return nil
}
```