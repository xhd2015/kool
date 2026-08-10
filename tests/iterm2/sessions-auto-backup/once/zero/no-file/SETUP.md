# Scenario

**Feature**: idle-only snapshot → 0 critical; no auto file created

```
Caller
  -> sessions auto-backup --once --file empty-auto.json
  -> idle-only fixture
  <- 0 critical; empty-auto.json not created
```

## Steps

1. ModeAutoBackup; Once; UseIdleOnlyFixture; FilePath=empty-auto.json.

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
	req.FilePath = "empty-auto.json"
	return nil
}
```
