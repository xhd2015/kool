# Scenario

**Feature**: `-n` / `--new-window` and `--no-new-window` are mutually exclusive

```
run "Say Hello" -n --no-new-window
  -> error exit ≠ 0; conflict message
```

## Steps

1. echoLeaf fixture (valid label); NewWindow and NoNewWindow both true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoLeafJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Say Hello"
	req.NewWindow = true
	req.NoNewWindow = true
	req.DryRun = false
	return nil
}
```
