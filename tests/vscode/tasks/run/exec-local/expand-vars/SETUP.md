# Scenario

**Feature**: real local run expands workspaceFolder / basename / env like dry-run

```
run "Echo Root" --backend=local
  with KOOL_TASKS_TEST_TOKEN=tok123
  -> stdout contains abs workspace path, basename, tok123
  -> no leftover ${workspaceFolder} / ${env:…}
```

## Steps

1. expandTaskJSONC; abs WorkingDir; Env token; Query=`Echo Root`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, expandTaskJSONC)
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
