# Scenario

**Feature**: list --json emits machine-readable JSON without ANSI

```
multi-task fixture + --json
  -> stdout JSON array/object of tasks; exit 0
```

## Steps

1. Multi-task fixture; JSON=true.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.JSON = true
	return nil
}
```
