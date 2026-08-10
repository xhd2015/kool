# Scenario

**Feature**: --check-update when codex missing reports missing; exit 0

```
LookPath miss
  -> kool codex install --check-update
  -> exit 0
  -> stdout status: missing (or not installed / not found)
  -> ShellCalls empty
  -> FetchLatestCalls == 0
```

## Steps

1. Present=false.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Present = false
	req.LocalRaw = ""
	req.Latest = ""
	return nil
}
```
