# Scenario

**Feature**: bare coverage without subcommand is an error

```
user -> HandleWith([])
  -> non-zero exit; stderr hints package-table or help
```

## Steps

1. Empty Args slice (no subcommand).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Explicit empty slice: structured builder would also yield empty args.
	req.Args = []string{}
	return nil
}
```
