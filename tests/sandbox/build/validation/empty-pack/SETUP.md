# Scenario

**Feature**: empty pack (no files and no env) is rejected

```
kool sandbox build -o sandbox.bin
  -> non-zero; stderr explains empty pack / no files / no env
```

## Steps

1. -o set; no -i; no --file; no --env.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	req.Input = ""
	req.InputSet = false
	req.ExtraFiles = nil
	req.ExtraEnv = nil
	return nil
}
```
