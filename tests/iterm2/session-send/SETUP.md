# Scenario

**Feature**: `kool iterm2 session send` flag form vs positional `session <id> send`

```text
session send (--session-id|--tab|--tab-index) <text> -> resolve -> SendText
session <id> send <text> -> resolve id -> SendText (unchanged)
```

## Preconditions

- In-process `RunForTestEnv` with injected ListSessions / CurrentStatus / SendText.
- Default fixture: two tabs in one window; current = tab 1.

## Steps

1. Leaf sets `Request.Args` (tokens after `session`).
2. Run invokes handler; Assert checks exit, stdout/stderr, SendCalls.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	if req.CurrentSessionID == "" {
		req.CurrentSessionID = fixtureTab1ID
	}
	return nil
}
```
