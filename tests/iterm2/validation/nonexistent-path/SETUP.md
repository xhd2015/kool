# Scenario

**Feature**: nonexistent directory path

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markValidationTree()
	markRootTree()
	req.DirPath = filepath.Join(req.WorkingDir, "no-such-dir")
	return nil
}
```