# Scenario

**Feature**: cloudflare help mode (no tunnel)

```
# user asks for help at root or serve
user -> kool cloudflare [--help | serve --help]
  -> handler prints usage, exit 0 (no StartSession)
```

## Steps

1. Mark help branch; StartSession must not be called (AllowStart remains false).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.AllowStart = false
	return nil
}
```
