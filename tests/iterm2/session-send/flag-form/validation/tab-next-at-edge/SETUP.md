# Scenario

**Feature**: `--tab next` on last tab fails (no wrap)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CurrentSessionID = fixtureTab2ID
	req.Args = []string{"send", "--tab", "next", "hi"}
	return nil
}
```
