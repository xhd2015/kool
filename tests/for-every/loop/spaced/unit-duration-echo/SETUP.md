# Scenario

**Feature**: spaced form with unit duration runs echo max-runs times

```
kool for-every --max-runs 2 10ms echo hello-spaced
  -> stdout: hello-spaced\nhello-spaced\n ; exit 0
```

## Steps

1. Duration 10ms; max-runs 2; echo one token.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Duration = "10ms"
	req.MaxRuns = intPtr(2)
	req.Command = "echo"
	req.Args = []string{"hello-spaced"}
	return nil
}
```
