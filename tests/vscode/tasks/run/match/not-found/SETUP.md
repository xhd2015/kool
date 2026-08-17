# Scenario

**Feature**: run unknown label errors

```
run "zzzz-missing" --dry-run -> Error not found
```

## Steps

1. Multi-task; Query missing.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "zzzz-missing"
	return nil
}
```
