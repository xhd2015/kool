# Scenario

**Feature**: --check-update when outdated reports update available; exit 0

```
LookPath hit + local 0.1.0 + latest 0.2.0
  -> kool codex install --check-update
  -> exit 0
  -> stdout status: update available (or equivalent) + versions
  -> ShellCalls empty
```

## Steps

1. Present; outdated versions.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Present = true
	req.LocalRaw = "codex-cli 0.1.0"
	req.Latest = "0.2.0"
	req.LatestFail = false
	return nil
}
```
