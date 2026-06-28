# Scenario

**Feature**: in-process `--ipc-only` does not URI-fallback when IPC fails

```
# IPC exhausted; ipc-only returns error
OpenDirOptions(IpcOnly) -> IPC fail -> error (no exec, no hint)
```

## Steps

1. Create valid directory.
2. Set `IPCAlwaysFail` so no mock server accepts connections.
3. Call `OpenDirOptions` with `IpcOnly: true`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "ipc-only-fail")
	req.DirPath = dir
	req.IpcOnly = true
	req.IPCAlwaysFail = true
	return nil
}
```