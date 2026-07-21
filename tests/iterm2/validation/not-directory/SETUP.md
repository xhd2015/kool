# Scenario

**Feature**: path exists but is not a directory

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markValidationTree()
	markRootTree()
	file := filepath.Join(req.WorkingDir, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		return err
	}
	req.DirPath = file
	return nil
}
```