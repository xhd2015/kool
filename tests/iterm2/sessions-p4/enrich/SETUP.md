# Scenario

**Feature**: sessions snapshot with injectable procresolve agent enrich

```
# busy pane on ttys002 + AgentResolveByTTY
sessions snapshot --no-color [--no-enrich|--no-tree|--json]
  -> process enrich (busy) then resolve attach
  -> CLI or JSON shows agent session / tree per flags
```

## Steps

1. Mode=snapshot-cli; two-window fixture; NoColor; install busy-tty resolve
   by default (leaves may override map or flags).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSnapshotCLI
	req.UseTwoWindowFixture = true
	req.NoColor = true
	if req.ITermRunning == nil {
		req.ITermRunning = boolPtr(true)
	}
	// Default resolve inject on busy tty; leaves may replace.
	if req.AgentResolveByTTY == nil {
		req.AgentResolveByTTY = map[string]iterm2.AgentResolveFixture{
			"ttys002": busyGrokResolve(),
		}
	}
	return nil
}
```
