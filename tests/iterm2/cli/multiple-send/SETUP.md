# Scenario

**Feature**: repeatable --send flags preserve order

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DirPath = initValidDir(t, req.WorkingDir, "send-multi")
	req.Send = []string{"grok", "codex"}
	return nil
}
```