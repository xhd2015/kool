# Scenario

**Feature**: local backend expands composite dependsOn and runs leaf steps

```
run "Both Steps" --backend=local
  -> Step One then Step Two (sequential local OK)
  -> stdout shows both KOOL_TASKS_P2_STEP_ONE and _STEP_TWO
```

## Steps

1. echoCompositeJSONC; Query=`Both Steps`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, echoCompositeJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Both Steps"
	return nil
}
```
