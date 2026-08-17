# Scenario

**Feature**: IPC retries after transient connect failure

```
# first accept rejects connection; second succeeds
OpenDir -> IPC connect (fail) -> retry -> IPC open -> ok:true
```

## Steps
1. Create valid directory.
2. Configure mock server to reject first connection.
3. Call `OpenDir`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dir := initValidDir(t, req.WorkingDir, "retry-target")
	req.DirPath = dir
	req.IPCFailConnects = 1
	return nil
}
```