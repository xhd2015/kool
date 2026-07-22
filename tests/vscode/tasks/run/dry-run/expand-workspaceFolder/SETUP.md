# Scenario

**Feature**: dry-run expands workspaceFolder, basename, and env vars

```
run "Echo Root" --dry-run
  with KOOL_TASKS_TEST_TOKEN=tok123
  -> plan shows absolute workspace path, basename, tok123
  -> no leftover ${workspaceFolder} / ${env:…}
```

## Steps

1. expandTaskJSONC fixture; Env sets token; Query=`Echo Root`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, expandTaskJSONC)
	// Abs path for stable plan checks
	abs, err := filepath.Abs(req.WorkingDir)
	if err != nil {
		return err
	}
	req.WorkingDir = abs
	req.Dir = abs
	req.Query = "Echo Root"
	req.Env = []string{"KOOL_TASKS_TEST_TOKEN=tok123"}
	return nil
}
```
