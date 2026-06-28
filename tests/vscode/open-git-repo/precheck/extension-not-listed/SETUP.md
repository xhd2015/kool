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
)

func Setup(t *testing.T, req *Request) error {
	repoDir := initValidGitRepo(t, req.WorkingDir, "repo")
	req.RepoPath = repoDir
	installNoExtensionPrecheck(t, req)
	return nil
}
```