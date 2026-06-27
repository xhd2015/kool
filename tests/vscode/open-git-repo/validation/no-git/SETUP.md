# Scenario

**Feature**: directory without `.git` fails with clear message

```
# plain directory lacks git metadata
validateGitRepoPath(plainDir) -> "not a git repository"
```

## Steps
1. Create a plain directory without `.git`.
2. Run CLI with that directory path.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Phase = "cli"
	dirPath := filepath.Join(req.WorkingDir, "plain-dir")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	req.RepoPath = dirPath
	return nil
}
```