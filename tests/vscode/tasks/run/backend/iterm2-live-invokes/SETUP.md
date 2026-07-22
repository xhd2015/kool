# Scenario

**Feature**: live iterm2 backend invokes RunTabSet (mock) for multi-leaf composite

```
run "Both Steps" --backend=iterm2
  + KOOL_VSCODE_TASKS_ITERM2_MOCK=1 / MOCK_OUT
  -> exit 0
  -> mock JSON once: 2 tabs (Step One, Step Two), windowName=Both Steps
  -> must NOT print "live iterm2 run not configured"
  -> must NOT fall back to local multi-echo of both markers as primary path
```

## Steps

1. echoCompositeJSONC; Backend=iterm2; DryRun=false; enableITerm2Mock.
2. Query=`Both Steps`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoCompositeJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Both Steps"
	req.Backend = "iterm2"
	req.DryRun = false
	enableITerm2Mock(t, req)
	return nil
}
```
