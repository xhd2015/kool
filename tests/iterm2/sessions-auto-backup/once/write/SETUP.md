# Scenario

**Feature**: --once with critical fixture writes auto checkpoint (version + sessions)

```
Caller
  -> sessions auto-backup --once --file sessions-auto.json
  -> critical fixture (grok + mark)
  <- Saved; JSON version=1; 2 sessions; restored_at empty
```

## Steps

1. ModeAutoBackup; Once; UseCriticalFixture; FilePath=sessions-auto.json.

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
	req.FilePath = "sessions-auto.json"
	return nil
}
```
