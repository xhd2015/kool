# Scenario

**Feature**: default enrich on busy grok pane shows session id + tree connectors

```
# fixture: busy ttys002 + shell→grok resolve (└──)
AgentResolveByTTY["ttys002"] = busyGrokResolve()
sessions snapshot --no-color
  -> stdout contains fixture grok session uuid
  -> stdout contains Unicode tree connector(s) └── and/or ├──
  -> optional title fixture-grok-title may appear
```

## Steps

1. Use default enrich/ inject (busyGrokResolve on ttys002); enrich on; tree on.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoEnrich = false
	req.NoTree = false
	// AgentResolveByTTY defaulted in enrich/SETUP.md to busyGrokResolve.
	return nil
}
```
