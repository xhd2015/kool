# Scenario

**Feature**: tab-set run with dry-run and flag validation

```
tab-set run <name> [--dry-run] [-n] [--no-new-window]
  -> dry-run plan | flag conflict error
```

## Steps

1. Subcommand `run`.
2. Leaves set DryRun / NewWindow / NoNewWindow and fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.Subcommand = "run"
	return nil
}
```
