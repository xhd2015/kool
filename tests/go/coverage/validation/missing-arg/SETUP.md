# Scenario

**Feature**: package-table without profile positional → usage exit 2

```
user -> HandleWith(["package-table"])
  -> exit 2; stderr usage/error about profile or arguments
```

## Steps

1. Subcommand package-table; no profile path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "package-table"
	// ProfilePath unset; ProfileSet false → no positional.
	return nil
}
```
