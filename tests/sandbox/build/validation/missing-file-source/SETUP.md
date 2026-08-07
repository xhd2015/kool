# Scenario

**Feature**: --file local path that does not exist is invalid

```
kool sandbox build -o sandbox.bin --file missing-local.txt=app/config.txt
  -> non-zero; stderr mentions file / not found / missing
```

## Steps

1. Point --file at a path that is not created under WorkingDir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	// Relative local path; deliberately not written by Setup.
	req.ExtraFiles = []string{"missing-local.txt=app/config.txt"}
	return nil
}
```
