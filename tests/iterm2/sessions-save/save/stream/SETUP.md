# Scenario

**Feature**: save dry-run progressive stream (per critical window)

```
Caller
  -> sessions save --dry-run
  -> ListTabs(1) -> classify -> write W1 …
  -> ListTabs(2)  # stdout already has W1 when this starts
  -> write W2 … -> footer Would save
```

## Steps

1. ModeSave; DryRun; two critical windows; observe stream order.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSave
	req.DryRun = true
	req.UseTwoCriticalWindows = true
	req.ObserveStreamOrder = true
	req.FilePath = "stream-plan.json"
	// No color flags: stream order is independent of --color/--no-color (W1 still matches).
	return nil
}
```
