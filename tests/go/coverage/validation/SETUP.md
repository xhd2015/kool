# Scenario

**Feature**: coverage routing and package-table arg validation fail fast

```
# bare / unknown
user -> coverage [no sub | nosuch]
  -> non-zero; stderr explains; no table

# missing profile arg
user -> package-table (no path)
  -> exit 2; usage/error

# missing file
user -> package-table /abs/missing.out
  -> exit 1; Error: on stderr
```

## Steps

1. Leaf sets Args or structured validation fields.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	// Validation branch: fail before a successful table; leaves set Args / flags.
	req.JSON = false
	return nil
}
```
