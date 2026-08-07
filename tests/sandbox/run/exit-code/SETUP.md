# Scenario

**Feature**: sealed runner propagates the guest process exit code

```
./sandbox.bin -- sh -c 'exit N'
  -> sealed binary exit code == N
```

## Steps

1. Exit-code leaves set guest shell to exit with a known status.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SealedDoubleDash = true
	return nil
}
```
