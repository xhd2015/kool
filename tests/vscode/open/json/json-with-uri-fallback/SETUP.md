# Scenario

**Feature**: `--json` reports URI fallback when IPC fails (default open mode)

```
# IPC fail then URI via exec hook; JSON on stdout
OpenDirOptions(Json, IpcOnly=false) -> IPC fail -> exec(vscode://...) -> JSON fallback
```

## Steps

1. Create valid directory.
2. Force IPC failure (`IPCAlwaysFail`).
3. Run with `Json: true` and **without** `IpcOnly`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "json-uri-fallback")
	req.DirPath = dir
	req.Json = true
	req.IpcOnly = false
	req.IPCAlwaysFail = true
	return nil
}
```