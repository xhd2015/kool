# Scenario

**Feature**: `--session-id` prefix resolve on flag form

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"send", "--session-id", "aaaaaaaa", "--no-submit", "hi"}
	return nil
}
```
