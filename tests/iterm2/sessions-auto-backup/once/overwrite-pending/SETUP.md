# Scenario

**Feature**: existing pending auto file is always overwritten (no TTY prompt/error)

```
Caller
  -> seed pending checkpoint (old-sess, restored_at null)
  -> sessions auto-backup --once --file pending-auto.json (non-TTY CI)
  -> critical fixture
  <- exit 0; file no longer old-sess; Saved
```

## Steps

1. ModeAutoBackup; Once; UseCriticalFixture; SeedDoc pending; FilePath=pending-auto.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeAutoBackup
	req.Once = true
	req.UseCriticalFixture = true
	req.FilePath = "pending-auto.json"
	req.SeedDoc = pendingSeedDoc()
	return nil
}
```
