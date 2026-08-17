# Scenario

**Feature**: missing directory argument

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DirPath = ""
	return nil
}
```