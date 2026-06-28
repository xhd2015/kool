# Scenario

**Feature**: `kool vscode open --ipc-only --json` reports IPC success on stdout

```
# subprocess with mock socket via KOOL_VSCODE_IPC_SOCKET
kool vscode open --ipc-only --json <dir> -> IPC {"op":"open"} -> ok:true
kool <- stdout {"ipc_handled":true,"path":"..."}
```

## Steps

1. Create valid directory under temp working dir.
2. Start mock IPC server on a temp Unix socket.
3. Set `req.IPCSocketPath` for CLI env override.
4. Run CLI with `--ipc-only` and `--json`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "ipc-json-ok")
	req.DirPath = dir
	req.IpcOnly = true
	req.Json = true
	socketPath := filepath.Join(req.WorkingDir, "mock-ipc.sock")
	startMockIPCServer(t, socketPath, 0)
	req.IPCSocketPath = socketPath
	return nil
}
```