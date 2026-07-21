# Scenario

**Feature**: file path (not directory) fails before open

```
# regular file is not a directory
validateGitRepoPath(file) -> error
```

## Steps
1. Create a regular file in the working directory.
2. Run CLI with that file path.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markValidationTree()
	markRootTree()
	req.Phase = "cli"
	filePath := filepath.Join(req.WorkingDir, "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		return err
	}
	req.RepoPath = filePath
	return nil
}
```