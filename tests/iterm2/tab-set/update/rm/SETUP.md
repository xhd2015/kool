# Scenario

**Feature**: update `--rm` removes a tab (confirm / --force; never empty tabs)

```
existing multi-tab set
  -> update <name> --tab-id <id> --rm [--force]
  -> tab removed (or error: last tab / non-TTY / exclusive flags)
```

## Steps

1. Leaves write multi-tab or single-tab fixtures.
2. Set Rm=true; success path uses Force=true (handler is non-TTY).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Rm = true
	return nil
}
```
