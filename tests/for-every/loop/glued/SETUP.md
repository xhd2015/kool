# Scenario

**Feature**: glued invocation `kool for-every-<duration> <command>…`

```
kool for-every-<duration> [OPTIONS] <command> [args...]
  -> duration from command suffix; same loop as spaced form
```

## Steps

1. Force glued form.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Glued = true
	return nil
}
```
