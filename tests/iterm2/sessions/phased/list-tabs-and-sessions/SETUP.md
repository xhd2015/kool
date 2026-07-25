# Scenario

**Feature**: ListTabsAndSessions(1) returns tabs and sessions for window 1

```
InstallPhasedFixtureCollectorForTest(...)
  -> c.ListTabsAndSessions(1)
  -> Tab-A with session AAAAAAAA-… on /dev/ttys001
```

## Steps

1. Mode=list-tabs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeListTabs
	return nil
}
```
