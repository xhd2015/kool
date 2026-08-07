# Scenario

**Feature**: build without -o/--output is invalid

```
kool sandbox build --env FOO=bar
  -> non-zero; stderr mentions output or -o
```

## Steps

1. Clear -o; provide non-empty env so failure is about missing output not empty pack.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Output = ""
	req.OutputSet = false
	req.ExtraEnv = []string{"FOO=bar"}
	return nil
}
```
