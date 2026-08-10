# Scenario

**Feature**: dry-run with current codex is noop / up to date

```
LookPath hit + local == latest (0.147.0)
  -> kool codex install --dry-run
  -> exit 0
  -> stdout dry-run + noop / up to date / current
  -> ShellCalls empty
```

## Steps

1. Present bin; local matches latest.

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
