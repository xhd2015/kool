# Scenario

**Feature**: live iterm2 `--no-new-window` selects no-new-window RunTabSet mode

```
run "Both Steps" --backend=iterm2 --no-new-window
  + mock
  -> mock mode = no-new-window
```

## Steps

1. echoCompositeJSONC; Backend=iterm2; NoNewWindow=true; enableITerm2Mock.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoCompositeJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Both Steps"
	req.Backend = "iterm2"
	req.DryRun = false
	req.NewWindow = false
	req.NoNewWindow = true
	enableITerm2Mock(t, req)
	return nil
}
```
