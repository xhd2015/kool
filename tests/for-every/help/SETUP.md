# Scenario

**Feature**: for-every help mode (no loop)

```
# user asks for help
user -> kool for-every --help
  -> handler prints usage, exit 0 (no duration parse, no loop)
```

## Steps

1. Fix Help=true for descendants.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Help = true
	return nil
}
```
