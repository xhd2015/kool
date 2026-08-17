# Scenario

**Feature**: run exact label match with --dry-run

```
run Compile --dry-run -> plan for Compile exit 0
```

## Steps

1. Multi-task; Query exact `Compile`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "Compile"
	return nil
}
```
