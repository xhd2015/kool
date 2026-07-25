# Scenario

**Feature**: streamed CLI window blocks include process idle/busy enrich

```
sessions snapshot --no-color
  -> fixture: ttys001 idle, ttys002 busy
  -> stdout contains both "idle" and "busy" labels for sessions
```

## Steps

1. Default stream; assert enrich labels on final stdout.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoStream = false
	return nil
}
```
