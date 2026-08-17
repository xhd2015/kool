# Scenario

**Feature**: spaced form missing command after duration

```
kool for-every --max-runs 1 10ms
  -> non-zero; requires command
```

## Steps

1. Spaced form (Glued=false).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Glued = false
	return nil
}
```
