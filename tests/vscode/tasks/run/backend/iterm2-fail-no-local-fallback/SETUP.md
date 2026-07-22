# Scenario

**Feature**: iterm2 live errors fail closed — no silent local multi-leaf fallback

```
run "Both Steps" --backend=iterm2
  + KOOL_VSCODE_TASKS_ITERM2_MOCK=1 + MOCK_ERR=1
  -> exit ≠ 0
  -> must NOT successfully local-run both leaves (both STEP markers)
  -> must NOT only warn and continue as local stub
```

## Steps

1. echoCompositeJSONC; Backend=iterm2; enableITerm2MockErr.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoCompositeJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Both Steps"
	req.Backend = "iterm2"
	req.DryRun = false
	enableITerm2MockErr(t, req)
	return nil
}
```
