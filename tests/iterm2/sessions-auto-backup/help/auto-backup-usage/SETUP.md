# Scenario

**Feature**: `sessions auto-backup -h` documents interval default 10m, --once, default file, core flags

```
Caller
  -> sessions auto-backup -h
  <- exit 0; mentions auto-backup, --interval, 10m, --once, sessions-auto.json, --file, --dry-run
```

## Steps

1. ModeHelp; HelpArgs = sessions auto-backup -h.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeHelp
	req.HelpArgs = []string{"sessions", "auto-backup", "-h"}
	return nil
}
```
