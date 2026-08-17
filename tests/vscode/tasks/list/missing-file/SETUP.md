# Scenario

**Feature**: list errors when no tasks.json found in walk-up

```
empty temp workspace (no .vscode/tasks.json)
  -> list -> Error exit 1
```

## Steps

1. WorkingDir is empty temp (root Setup); no tasks.json written.
2. Optional --dir = WorkingDir.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Dir = req.WorkingDir
	return nil
}
```
