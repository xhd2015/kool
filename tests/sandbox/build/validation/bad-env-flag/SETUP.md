# Scenario

**Feature**: --env without KEY=VALUE form is invalid

```
kool sandbox build -o sandbox.bin --env NOTVALID
  -> non-zero; stderr mentions env or =
```

## Steps

1. Pass malformed --env (no `=`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	req.ExtraEnv = []string{"NOTVALID"}
	return nil
}
```
