# Scenario

**Feature**: package-table --help documents flags and profile arg

```
user -> HandleWith(["package-table", "--help"])
  -> stdout documents --module --dir --skip-prefix --skip-contains --all --json
  -> exit 0; trailing \n
```

## Steps

1. Request package-table help.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HelpPackageTable = true
	return nil
}
```
