# Scenario

**Feature**: dry-run single leaf shell task prints command plan

```
run Compile --dry-run -> exit 0; plan mentions go build / Compile
```

## Steps

1. Multi-task fixture; Query=Compile; DryRun from parent.

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
