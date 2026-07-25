# Scenario

**Feature**: CaptureSnapshot uses phased ListWindows + per-window ListTabs

```
InstallPhasedFixtureCollectorForTest(OnListWindows, OnListTabs counters)
  -> CaptureSnapshot()
  -> ListWindowsCalls == 1
  -> ListTabsCalls == 2 (once per window)
  -> Summary.Windows == 2, Sessions == 2
```

## Steps

1. Mode=capture-phased.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeCapturePhased
	return nil
}
```
