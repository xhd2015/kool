# Scenario

**Feature**: dry-run composite expands dependsOn into plan steps

```
run "Build All" --dry-run
  -> plan includes Compile and Serve leaf steps
```

## Steps

1. Multi-task fixture; Query=`Build All`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "Build All"
	return nil
}
```
