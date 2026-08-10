# Scenario

**Feature**: invalid --interval → Error + non-zero before loop

```
Caller
  -> sessions auto-backup --once --interval not-a-duration
  <- Error: on stderr; non-zero exit (no hang)
```

## Steps

1. ModeAutoBackup; Once; Interval=not-a-duration (no fixture needed).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeAutoBackup
	req.Once = true
	req.Interval = "not-a-duration"
	// Validation should fail before capture; no fixture required.
	req.FilePath = "never-written.json"
	return nil
}
```
