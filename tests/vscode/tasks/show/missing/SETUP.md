# Scenario

**Feature**: show unknown label errors

```
show "No Such Task" -> Error exit ≠ 0
```

## Steps

1. Multi-task fixture; Query missing.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "No Such Task"
	return nil
}
```
