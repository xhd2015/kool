# Scenario

**Feature**: --dry-run lists existing sock basenames under events/

```
touch ROOT/events/sess-dry.sock
kool sandbox notify-event --type devbox.updated --path /abs/x --root ROOT --dry-run
  -> exit 0; stdout mentions sess-dry.sock (or path containing it)
```

## Steps

1. EnsureEventsDir + FakeSockNames=["sess-dry.sock"].
2. DryRun=true.
3. EventPath dummy abs.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.EnsureEventsDir = true
	req.FakeSockNames = []string{"sess-dry.sock"}
	req.MockListener = false
	req.DryRun = true
	req.EventType = "devbox.updated"
	req.EventPath = filepath.Join(req.WorkingDir, "some-load.bin")
	return nil
}
```
