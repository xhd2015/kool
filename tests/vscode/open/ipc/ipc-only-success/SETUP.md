# Scenario

**Feature**: in-process `--ipc-only` open succeeds without OS fallback

```
# OpenDirOptions.IpcOnly with mock IPC
OpenDirOptions(IpcOnly) -> IPC ok -> no exec hook
```

## Steps

1. Create valid directory.
2. Start mock IPC server.
3. Call `OpenDirOptions` with `IpcOnly: true`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	markIpcTree()
	markRootTree()
	dir := initValidDir(t, req.WorkingDir, "ipc-only-ok")
	req.DirPath = dir
	req.IpcOnly = true
	req.IPCFailConnects = 0
	return nil
}
```