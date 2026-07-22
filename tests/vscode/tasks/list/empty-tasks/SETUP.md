# Scenario

**Feature**: list with empty tasks array succeeds with zero count

```
tasks.json version 2, tasks: []
  -> list exit 0; count 0
```

## Steps

1. Write empty tasks fixture; --dir = WorkingDir.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeEmptyTasks(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	return nil
}
```
