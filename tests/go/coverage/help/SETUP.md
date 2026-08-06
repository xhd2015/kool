# Scenario

**Feature**: coverage help surfaces (root and package-table)

```
# root help
user -> kool go coverage --help
  -> usage lists package-table; exit 0; trailing newline

# package-table help
user -> kool go coverage package-table --help
  -> documents profile arg and filter/format flags; exit 0
```

## Steps

1. Leaf sets HelpAtRoot or HelpPackageTable.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	// Help branch: leaves set HelpAtRoot or HelpPackageTable; no profile work.
	req.ProfileSet = false
	return nil
}
```
