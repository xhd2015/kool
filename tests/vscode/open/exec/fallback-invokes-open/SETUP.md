# Scenario

**Feature**: `OpenDir` orchestration invokes OS `open` with built URI after IPC fail

```
# mocked exec captures open <uri> after IPC exhaustion
OpenDir(validDir) -> IPC fail -> exec("open", uri)
```

## Steps
1. Create valid directory.
2. Call `OpenDir` with IPC unavailable and mocked exec on darwin.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "exec-target")
	req.DirPath = dir
	return nil
}
```