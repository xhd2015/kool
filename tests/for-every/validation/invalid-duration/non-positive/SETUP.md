# Scenario

**Feature**: duration ≤ 0 is rejected

```
kool for-every 0s --max-runs 1 true
  -> non-zero; duration must be greater than 0
```

## Steps

1. Duration `0s` (parses but fails > 0 check).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Duration = "0s"
	return nil
}
```
