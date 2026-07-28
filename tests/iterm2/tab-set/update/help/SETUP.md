# Scenario

**Feature**: `tab-set update --help` documents key update flags

```
tab-set update --help
  -> exit 0; mentions update, --tab-id, --rm, --no-submit
```

## Steps

1. Subcommand=update (inherited); Help=true.
2. No SetName / no config fixture required.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Help = true
	return nil
}
```
