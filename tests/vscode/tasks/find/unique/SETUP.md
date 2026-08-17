# Scenario

**Feature**: find with unique case-insensitive match

```
query "compile" (CI) matches "Compile" only -> exit 0
```

## Steps

1. Multi-task fixture; Query=`compile`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "compile"
	return nil
}
```
