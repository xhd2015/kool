# Scenario

**Feature**: ensure current bin is a noop (no shell)

```
LookPath hit + local == latest (0.147.0)
  -> kool codex install
  -> exit 0
  -> ShellCalls empty
  -> FetchLatestCalls >= 1
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
