# Scenario

**Feature**: default `auto` backend runs one non-background leaf via local

```
run "Say Hello"   # no --backend → auto; one fg shell leaf
  -> local path; exit 0; KOOL_TASKS_P2_HELLO on stdout
  -> no live iTerm
```

## Steps

1. echoLeafJSONC; Query=`Say Hello`; Backend empty (product default auto); DryRun=false.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoLeafJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Say Hello"
	req.Backend = "" // product default: auto
	req.DryRun = false
	return nil
}
```
