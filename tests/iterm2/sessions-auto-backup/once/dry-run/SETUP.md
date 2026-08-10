# Scenario

**Feature**: --dry-run alone plans critical save, exits after one cycle (no loop); no file

```
Caller
  -> sessions auto-backup --dry-run --file plan-auto.json
  -> critical fixture
  <- Would save + session ids; plan-auto.json not created; process exits (no --once needed)
```

## Steps

1. ModeAutoBackup; DryRun only (Once=false); UseCriticalFixture; FilePath=plan-auto.json.
   Proves --dry-run implies single cycle exit without --once.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeAutoBackup
	req.Once = false
	req.DryRun = true
	req.UseCriticalFixture = true
	req.FilePath = "plan-auto.json"
	return nil
}
```
