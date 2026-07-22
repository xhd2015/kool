# Scenario

**Feature**: live iterm2 default window mode is smart

```
run "Both Steps" --backend=iterm2   # no -n / --no-new-window
  + mock
  -> mock mode = smart (or stdout mentions mode: smart)
```

## Steps

1. echoCompositeJSONC; Backend=iterm2; enableITerm2Mock; no window flags.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoCompositeJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Both Steps"
	req.Backend = "iterm2"
	req.DryRun = false
	req.NewWindow = false
	req.NoNewWindow = false
	enableITerm2Mock(t, req)
	return nil
}
```
