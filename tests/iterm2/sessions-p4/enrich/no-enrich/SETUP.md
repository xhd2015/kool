# Scenario

**Feature**: --no-enrich skips procresolve attach entirely

```
# same busy-grok inject installed, but flag suppresses output
sessions snapshot --no-color --no-enrich
  -> exit 0; busy/idle still shown
  -> must NOT contain fixture agent session uuid
  -> must NOT contain FormatTree connectors from agent tree
```

## Steps

1. Keep busyGrokResolve inject; set NoEnrich=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoEnrich = true
	req.NoTree = false
	return nil
}
```
