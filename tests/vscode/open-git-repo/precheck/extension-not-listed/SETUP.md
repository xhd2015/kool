# Scenario

**Feature**: unlisted extension blocks open-git-repo

```
# code present but extension missing from --list-extensions
EnsureExtensionListed -> error (extension id + install hint)
```

## Steps
1. Create valid git repo.
2. Install fake `code` listing other extensions only.
3. Run CLI.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repoDir := initValidGitRepo(t, req.WorkingDir, "repo")
	req.RepoPath = repoDir
	installNoExtensionPrecheck(t, req)
	return nil
}
```