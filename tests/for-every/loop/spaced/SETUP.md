# Scenario

**Feature**: spaced invocation `kool for-every <duration> <command>…`

```
kool for-every [OPTIONS] <duration> <command> [args...]
  -> parse duration positional, then run loop
```

## Steps

1. Force spaced form (Glued=false).

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
