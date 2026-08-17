# Scenario

**Feature**: cloudflare validation errors before StartSession

```
# invalid argv → no tunnel start
user -> kool cloudflare … (no subcommand | unknown | missing flags)
  -> stderr error, non-zero exit; StartSession not called
```

## Steps

1. Mark validation-only: AllowStart stays false.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.AllowStart = false
	req.HelpAtRoot = false
	req.HelpServe = false
	return nil
}
```
