# Scenario

**Feature**: dry-run with outdated codex plans update (UpdateCmd) + versions

```
LookPath hit + local 0.1.0 + latest 0.2.0
  -> kool codex install --dry-run
  -> exit 0
  -> stdout dry-run + would update
  -> versions 0.1.0 and 0.2.0
  -> UpdateCmd present
  -> ShellCalls empty
```

## Steps

1. Present bin; local older than latest.

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
