# Scenario

**Feature**: `-n` / `--new-window` and `--no-new-window` are mutually exclusive

```
run "Say Hello" -n --no-new-window
  -> error exit ≠ 0; conflict message
```

## Steps

1. echoLeaf fixture (valid label); NewWindow and NoNewWindow both true.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeTasksJSON(t, req.WorkingDir, echoLeafJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Say Hello"
	req.NewWindow = true
	req.NoNewWindow = true
	req.DryRun = false
	return nil
}
```
