# Scenario

**Feature**: capture fail with --once → soft warning + exit 0; no file written

```
Caller
  -> sessions auto-backup --once --file fail-auto.json
  -> FailSnapshotCapture (iTerm not running)
  <- warning: on stderr; exit 0; file not created
```

## Steps

1. ModeAutoBackup; Once; FailSnapshotCapture; FilePath=fail-auto.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeAutoBackup
	req.Once = true
	req.FailSnapshotCapture = true
	req.FilePath = "fail-auto.json"
	return nil
}
```
