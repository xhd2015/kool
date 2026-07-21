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
)

func Setup(t *testing.T, req *Request) error {
	markPrecheckTree()
	markRootTree()
	repoDir := initValidGitRepo(t, req.WorkingDir, "repo")
	req.RepoPath = repoDir
	req.CodeInPath = false
	return nil
}
```