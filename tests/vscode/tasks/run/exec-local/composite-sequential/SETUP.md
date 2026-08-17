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
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeTasksJSON(t, req.WorkingDir, echoCompositeJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Both Steps"
	return nil
}
```
