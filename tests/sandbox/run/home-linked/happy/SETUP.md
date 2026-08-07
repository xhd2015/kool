# Scenario

**Feature**: home-linked happy paths (HOME, seed, overlay, explode)

```
# happy home-linked run
HOME=FAKE KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- <cmd>
  -> exit 0; guest sees seeded/overlaid tree and HOME=session root
```

## Steps

1. Happy leaves pack files/fake-home as needed and set `SealedArgs`.
2. Prefer double-dash before guest argv (same as run/happy).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// End runner flags with -- before guest argv (portable sh -c one-liners).
	req.SealedDoubleDash = true
	return nil
}
```
