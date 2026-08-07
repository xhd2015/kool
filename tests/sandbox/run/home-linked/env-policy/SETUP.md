# Scenario

**Feature**: packed HOME policy under home-linked (conflict errors)

```
# packed HOME must not conflict with home-linked guest HOME
kool sandbox build -o sandbox.bin --home-linked --env HOME=/tmp/not-sandbox …
HOME=FAKE KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- true
  -> non-zero; stderr mentions HOME and/or home-linked
```

## Steps

1. Policy leaves pack conflicting HOME; ensure sealed run is enabled without
   double-dash unless a leaf sets it (error path before guest).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AfterBuildRun = true
	req.SealedDoubleDash = false
	return nil
}
```
