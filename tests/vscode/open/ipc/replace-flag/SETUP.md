# Scenario

**Feature**: IPC open with `--replace` sends `replace: true`

```
# replace propagates into IPC JSON
OpenDir(replace=true) -> IPC {"op":"open","path":"/abs/dir","replace":true} -> ok
```

## Steps
1. Create valid directory.
2. Start mock IPC server.
3. Call `OpenDir` with `replace=true`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "ipc-replace-target")
	req.DirPath = dir
	req.Replace = true
	req.IPCFailConnects = 0
	return nil
}
```