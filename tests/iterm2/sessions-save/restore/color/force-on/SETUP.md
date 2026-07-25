# Scenario

**Feature**: `restore --dry-run --color` forces ANSI on plan output

```
Caller
  -> sessions restore --dry-run --color --file restore-color.json
  <- stdout contains ESC CSI; green Would restore and/or bold new window
```

## Steps

1. Color=true.

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
