# Scenario

**Feature**: live iterm2 `-n` selects new-window RunTabSet mode

```
run "Both Steps" --backend=iterm2 -n
  + mock
  -> mock mode = new-window
```

## Steps

1. echoCompositeJSONC; Backend=iterm2; NewWindow=true; enableITerm2Mock.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoCompositeJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Both Steps"
	req.Backend = "iterm2"
	req.DryRun = false
	req.NewWindow = true
	req.NoNewWindow = false
	enableITerm2Mock(t, req)
	return nil
}
```
