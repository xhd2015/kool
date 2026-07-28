# Scenario

**Feature**: tab-set list reads config directory

```
KOOL_ITERM2_TAB_SET_DIR -> tab-set list -> set names on stdout
```

## Steps

1. Subcommand `list`.
2. Leaves prepare empty dir or fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.Subcommand = "list"
	return nil
}
```
