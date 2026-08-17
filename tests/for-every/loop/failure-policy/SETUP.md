# Scenario

**Feature**: child failure handling and consecutive-failure counters

```
# default: continue on failure (until max-runs / signal)
# --allow-failure alone ≡ max-failure 1 (exit on first failure)
# --max-failure N stops after N consecutive failures
# success resets consecutive counter; both flags → max-failure wins
```

## Steps

1. Short interval; leaves set stop flags and a failing (or mixed) child.
2. Prefer spaced form for failure-policy leaves (form is not under test here).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Glued = false
	req.Duration = "10ms"
	return nil
}
```
