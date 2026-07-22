# Scenario

**Feature**: list walks up from nested --dir to parent tasks.json

```
workspace/.vscode/tasks.json + nested/ subdir
  -> list --dir nested -> finds parent workspace tasks
```

## Steps

1. Write multi-task at workspace root.
2. Create nested child dir; set Dir to nested path (or WorkingDir=nested without --dir).

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	nested := filepath.Join(req.WorkingDir, "pkg", "nested")
	if err := os.MkdirAll(nested, 0755); err != nil {
		return err
	}
	// Start discovery from nested dir via --dir
	req.Dir = nested
	return nil
}
```
