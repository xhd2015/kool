# Scenario

**Feature**: IPC success opens directory without OS fallback

```
# IPC responds ok; exec hook not invoked
OpenDir -> IPC {"op":"open"} -> ok:true
```

## Steps
1. Create valid directory.
2. Start mock IPC server.
3. Call `OpenDir` with exec hook installed.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markIpcTree()
	markRootTree()
	dir := initValidDir(t, req.WorkingDir, "ipc-target")
	req.DirPath = dir
	req.IPCFailConnects = 0
	return nil
}
```