# Scenario

**Feature**: single --send follow-up

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DirPath = initValidDir(t, req.WorkingDir, "send-one")
	req.Send = []string{"grok"}
	return nil
}
```