# Scenario

**Feature**: `v0.0.87` increments to `v0.0.88`

```
IncrementTag("v0.0.87") -> "v0.0.88"
```

## Steps

1. Set `req.Tag` to `v0.0.87`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Tag = "v0.0.87"
	return nil
}
```