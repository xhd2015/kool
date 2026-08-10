# Scenario

**Feature**: publisher delivers devbox.updated JSON to a fake unix listener

```
listen ROOT/events/sess-mock.sock
kool sandbox notify-event --type devbox.updated --path ABS_LOAD --root ROOT
  -> DeliveredCount >= 1
  -> JSON: v=1, type=devbox.updated, path=ABS_LOAD, ts non-empty RFC3339-ish
  -> exit 0; stdout success summary (counts optional wording)
```

## Steps

1. MockListener with default sock name `sess-mock.sock`.
2. EventPath = absolute load path string (file need not exist for deliver framing).
3. DryRun=false.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.EnsureEventsDir = true
	req.MockListener = true
	req.MockSockName = "sess-mock.sock"
	req.DryRun = false
	req.EventType = "devbox.updated"
	req.EventPath = filepath.Join(req.WorkingDir, "load-target.bin")
	return nil
}
```
