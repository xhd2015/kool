# Scenario

**Feature**: session status shows runner (grok) session from inject resolve

```
session BBBBBBBB-0000-0000-0000-000000000002 status --no-color
  + AgentResolveByTTY ttys002 = busyGrokResolve
  -> exit 0
  -> stdout contains fixture grok session uuid
  -> stdout contains grok
  -> preferably tree connector └── (status parity with snapshot)
```

## Steps

1. Default status/ inject; enrich + tree on.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoEnrich = false
	req.NoTree = false
	req.SessionRef = fixtureITermBusyID
	return nil
}
```
