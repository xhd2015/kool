# Scenario

**Feature**: update validation and exclusive-flag errors

```
update … with bad args / missing target
  -> exit ≠ 0; config file unchanged when present
```

## Steps

1. Leaves set the conflicting or incomplete flags.
2. Most leaves use bots fixture except missing-set.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Leaves configure error scenarios.
	return nil
}
```
