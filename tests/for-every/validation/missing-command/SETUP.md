# Scenario

**Feature**: duration present but child command missing

```
# spaced or glued with valid duration and no command token
user -> kool for-every[-<dur>] [OPTIONS] <dur?>
  -> non-zero validation; no loop
```

## Steps

1. Provide a valid short duration; leave Command empty; pass max-runs for safety.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Duration = "10ms"
	req.Command = ""
	req.MaxRuns = intPtr(1)
	return nil
}
```
