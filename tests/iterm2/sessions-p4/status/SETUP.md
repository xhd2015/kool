# Scenario

**Feature**: session status uses the same procresolve enrich as snapshot

```
# single-session status
kool iterm2 session <busy-iterm-id> status --no-color
  -> Capture + find session + resolve attach
  -> CLI shows runner agent session id
```

## Steps

1. Mode=session-status; two-window fixture; NoColor; resolve on busy tty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSessionStatus
	req.UseTwoWindowFixture = true
	req.NoColor = true
	if req.ITermRunning == nil {
		req.ITermRunning = boolPtr(true)
	}
	if req.SessionRef == "" {
		req.SessionRef = fixtureITermBusyID
	}
	if req.AgentResolveByTTY == nil {
		req.AgentResolveByTTY = map[string]iterm2.AgentResolveFixture{
			"ttys002": busyGrokResolve(),
		}
	}
	return nil
}
```
