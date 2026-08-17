# Scenario

**Feature**: glued form missing command after flags

```
kool for-every-10ms --max-runs 1
  -> non-zero; requires command
```

## Steps

1. Glued form with duration suffix only.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Glued = true
	return nil
}
```
