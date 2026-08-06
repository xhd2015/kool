# Scenario

**Feature**: coverage root --help lists package-table

```
user -> HandleWith(["--help"])
  -> stdout usage mentions package-table; exit 0; ends with \n
```

## Steps

1. Request root help only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HelpAtRoot = true
	return nil
}
```
