# Scenario

**Feature**: capture fail must not clobber existing auto backup

```
Caller
  -> seed pending auto file (old-sess)
  -> sessions auto-backup --once --file keep-fail.json
  -> FailSnapshotCapture
  <- warning:; exit 0; file still old-sess
```

## Steps

1. ModeAutoBackup; Once; FailSnapshotCapture; SeedDoc pending; FilePath=keep-fail.json.

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
	req.FilePath = "keep-fail.json"
	req.SeedDoc = pendingSeedDoc()
	return nil
}
```
