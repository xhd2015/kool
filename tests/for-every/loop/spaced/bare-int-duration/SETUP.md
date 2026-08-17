# Scenario

**Feature**: bare integer duration is seconds (shared with timeout)

```
kool for-every --max-runs 1 1 true
  -> one successful run; exit 0 (duration "1" ≡ 1s)
```

## Steps

1. Duration `1` (bare int); max-runs 1; command `true`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Duration = "1"
	req.MaxRuns = intPtr(1)
	req.Command = "true"
	req.Args = nil
	return nil
}
```
