# Scenario

**Feature**: dry-run detects dependsOn cycles

```
Cycle A <-> Cycle B -> run "Cycle A" --dry-run -> Error cycle
```

## Steps

1. cycleJSONC; Query=`Cycle A`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, cycleJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Cycle A"
	return nil
}
```
