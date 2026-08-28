# Scenario

**Feature**: positional form rejects `--tab`

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"11111111", "send", "--tab", "next", "hi"}
	return nil
}
```
