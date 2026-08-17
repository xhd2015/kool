# Scenario

**Feature**: `kool vscode open --ipc-only --json` reports IPC failure without URI fallback

```
# IPC unreachable; ipc-only blocks OS opener
kool vscode open --ipc-only --json <dir> -> IPC (fail)
kool <- stdout {"ipc_handled":false,"path":"...","error":"..."}
```

## Steps

1. Create valid directory.
2. Point socket env at a path with **no** listening server (IPC always fails).
3. Run CLI with `--ipc-only` and `--json`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dir := initValidDir(t, req.WorkingDir, "ipc-json-fail")
	req.DirPath = dir
	req.IpcOnly = true
	req.Json = true
	req.IPCSocketPath = filepath.Join(req.WorkingDir, "missing-ipc.sock")
	return nil
}
```