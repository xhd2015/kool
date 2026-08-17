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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dir := initValidDir(t, req.WorkingDir, "ipc-target")
	req.DirPath = dir
	req.IPCFailConnects = 0
	return nil
}
```