# Scenario

**Feature**: --once --dry-run plans critical save; no file written

```
Caller
  -> sessions auto-backup --once --dry-run --file plan-auto.json
  -> critical fixture
  <- Would save + session ids; plan-auto.json not created
```

## Steps

1. ModeAutoBackup; Once; DryRun; UseCriticalFixture; FilePath=plan-auto.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeAutoBackup
	req.Once = true
	req.DryRun = true
	req.UseCriticalFixture = true
	req.FilePath = "plan-auto.json"
	return nil
}
```
