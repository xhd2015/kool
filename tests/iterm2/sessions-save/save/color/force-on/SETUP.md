# Scenario

**Feature**: `save --dry-run --color` forces ANSI on plan output

```
Caller
  -> sessions save --dry-run --color --file color-plan.json
  -> critical fixture
  <- stdout contains ESC CSI; green verb and/or bold W present
```

## Steps

1. Color=true (force on).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Color = true
	return nil
}
```
