# Scenario

**Feature**: `save --dry-run --no-color` suppresses all ANSI

```
Caller
  -> sessions save --dry-run --no-color
  -> critical fixture
  <- stdout has no \x1b; still Would save + critical content
```

## Steps

1. NoColor=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoColor = true
	return nil
}
```
