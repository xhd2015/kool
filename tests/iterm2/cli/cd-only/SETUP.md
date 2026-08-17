# Scenario

**Feature**: cd-only open without follow-ups

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DirPath = initValidDir(t, req.WorkingDir, "cd-only-target")
	return nil
}
```