# Scenario

**Feature**: glued for-every-10ms echoes max-runs times

```
kool for-every-10ms --max-runs 2 echo hello-glued
  -> stdout: hello-glued\nhello-glued\n ; exit 0
```

## Steps

1. Duration suffix 10ms; max-runs 2; echo.

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
	req.Args = []string{"hello-glued"}
	return nil
}
```
