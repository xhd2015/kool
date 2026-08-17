# Scenario

**Feature**: run --dry-run expands plan without process spawn

```
run <label> --dry-run
  -> plan workspace + steps; no child process / iTerm
```

## Steps

1. DryRun=true for all descendants.
2. Leaves set Query and specialized fixtures.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	return nil
}
```
