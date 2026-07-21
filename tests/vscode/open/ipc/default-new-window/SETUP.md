# Scenario

**Feature**: default IPC open omits `replace` field (new window)

```
# IPC client sends open without replace key
OpenDir(replace=false) -> IPC {"op":"open","path":"/abs/dir"} -> ok
```

## Steps
1. Create valid directory.
2. Start mock IPC server.
3. Call `OpenDir` without `--replace`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markIpcTree()
	markRootTree()
	dir := initValidDir(t, req.WorkingDir, "ipc-default-target")
	req.DirPath = dir
	req.Replace = false
	req.IPCFailConnects = 0
	return nil
}
```