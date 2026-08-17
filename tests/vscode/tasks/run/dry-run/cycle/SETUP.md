# Scenario

**Feature**: dry-run detects dependsOn cycles

```
Cycle A <-> Cycle B -> run "Cycle A" --dry-run -> Error cycle
```

## Steps

1. cycleJSONC; Query=`Cycle A`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeTasksJSON(t, req.WorkingDir, cycleJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Cycle A"
	return nil
}
```
