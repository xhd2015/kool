# Scenario

**Feature**: extra positional arguments after directory

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DirPath = initValidDir(t, req.WorkingDir, "proj")
	req.ExtraPositional = []string{"extra"}
	return nil
}
```