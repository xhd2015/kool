# Scenario

**Feature**: `--tab-index 0` targets the first tab (0-based)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"send", "--tab-index", "0", "hi"}
	return nil
}
```
