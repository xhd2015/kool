# Scenario

**Feature**: file path fails before precheck

```
# path exists but is a regular file
ValidateDirPath(file) -> not a directory error
```

## Steps
1. Create a file under working dir.
2. Run CLI with file path.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markValidationTree()
	markRootTree()
	filePath := filepath.Join(req.WorkingDir, "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		return err
	}
	req.DirPath = filePath
	return nil
}
```