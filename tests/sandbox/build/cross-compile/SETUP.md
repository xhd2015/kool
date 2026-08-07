# Scenario

**Feature**: cross-compile sealed binary with --goos / --goarch

```
user -> kool sandbox build -o OUT --goos OS --goarch ARCH …
  -> OUT is a binary for the requested target
```

## Steps

1. Leaves set Goos/Goarch and a non-empty pack.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	req.BuildTwice = false
	req.AfterBuildInspect = false
	return nil
}
```
