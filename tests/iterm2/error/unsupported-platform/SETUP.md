# Scenario

**Feature**: non-darwin platform rejected

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Phase = "handler"
	req.GoOS = "linux"
	req.DirPath = initValidDir(t, req.WorkingDir, "proj")
	return nil
}
```