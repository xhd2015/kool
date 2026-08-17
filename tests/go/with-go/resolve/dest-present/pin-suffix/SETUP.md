# Scenario

**Feature**: naked `1.19` pins to dest suffix `go1.19.13`

```
# dest-exists fixture; PinPatch(1.19) == go1.19.13
1.19 -> ResolveGorootWith -> path ends with go1.19.13
```

## Steps

1. Override `GoVersion` to naked `1.19`.
2. Dest dir `$InstallDir/go1.19.13` already created by dest-present Setup.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.GoVersion = "1.19"
	return nil
}
```
