# Scenario

**Feature**: `v0.2.1` increments to `v0.2.2`

```
IncrementTag("v0.2.1") -> "v0.2.2"
```

## Steps

1. Set `req.Tag` to `v0.2.1`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Tag = "v0.2.1"
	return nil
}
```