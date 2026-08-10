# Scenario

**Feature**: --check-update when current reports up to date; exit 0

```
LookPath hit + local == latest (0.147.0)
  -> kool codex install --check-update
  -> exit 0
  -> stdout status: up to date
  -> ShellCalls empty
```

## Steps

1. Present; current versions.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Present = true
	req.LocalRaw = "codex-cli 0.147.0"
	req.Latest = "0.147.0"
	req.LatestFail = false
	return nil
}
```
