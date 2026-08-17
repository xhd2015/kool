# Scenario

**Feature**: run label matching — exact, unique substring, ambiguous, not found

```
run <query> --dry-run
  exact first; else unique CI substring; else error
```

## Steps

1. DryRun=true for match leaves (plan path only).
2. Leaves set Query and fixtures.

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
