# Scenario

**Feature**: empty events dir → warning + exit 0 (no subscribers)

```
mkdir -p ROOT/events   # empty
kool sandbox notify-event --type devbox.updated --path /abs/load.bin --root ROOT
  -> exit 0; stderr warns no subscribers / no socks; stdout optional summary
```

## Steps

1. EnsureEventsDir=true (empty events/).
2. EventPath = abs dummy path under WorkingDir (need not exist for publisher).
3. RunNotifyEvent (inherited).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.EnsureEventsDir = true
	req.FakeSockNames = nil
	req.MockListener = false
	req.DryRun = false
	// Absolute path for --path; publisher must not require the file for no-socks.
	req.EventPath = filepath.Join(req.WorkingDir, "missing-load.bin")
	req.EventType = "devbox.updated"
	return nil
}
```
