# Scenario

**Feature**: IPC failure triggers stderr hint and vscode:// URI fallback

```
# no IPC server; retries exhausted
OpenDir -> IPC (fail) -> stderr hint -> OS opener(vscode:// URI)
```

## Steps
1. Create valid directory.
2. Do not start IPC server (`IPCAlwaysFail`).
3. Call `OpenDir` on darwin with exec hook.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "fallback-target")
	req.DirPath = dir
	req.IPCAlwaysFail = true
	return nil
}
```