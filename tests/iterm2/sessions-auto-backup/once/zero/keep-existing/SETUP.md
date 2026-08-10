# Scenario

**Feature**: 0 critical must not clobber existing auto backup

```
Caller
  -> seed pending auto file (old-sess)
  -> sessions auto-backup --once --file keep-auto.json
  -> idle-only fixture
  <- 0 critical; file still old-sess
```

## Steps

1. ModeAutoBackup; Once; UseIdleOnlyFixture; SeedDoc pending; FilePath=keep-auto.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeAutoBackup
	req.Once = true
	req.UseIdleOnlyFixture = true
	req.FilePath = "keep-auto.json"
	req.SeedDoc = pendingSeedDoc()
	return nil
}
```
