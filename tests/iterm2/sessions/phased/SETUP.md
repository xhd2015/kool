# Scenario

**Feature**: phased hierarchy APIs on SnapshotCollector (no CLI stream)

```
# composable AppleScript phases
collector.ListWindows() -> window headers
collector.ListTabsAndSessions(windowIndex) -> tabs + sessions
CaptureSnapshot() -> ListWindows + per-window ListTabs + process enrich
```

## Steps

1. Mark phased branch; install fixture via leaf Mode.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Phased leaves call package APIs with fixture inject (not CLI).
	req.UseTwoWindowFixture = true
	if req.ITermRunning == nil {
		req.ITermRunning = boolPtr(true)
	}
	return nil
}
```
