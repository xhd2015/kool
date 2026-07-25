# Scenario

**Feature**: --no-stream buffers full CLI (no progressive W1 during last ListTabs)

```
sessions snapshot --no-stream --no-color
  -> collect all windows first
  -> when ListTabs(2) starts, stdout does NOT yet contain W1
  -> final stdout still full CLI with W1, W2, footer
```

## Steps

1. Set NoStream=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoStream = true
	return nil
}
```
