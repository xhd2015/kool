# Scenario

**Feature**: show shell leaf prints command and options

```
show Compile -> type shell, command go build, cwd workspaceFolder
```

## Steps

1. Multi-task fixture; Query=`Compile`.

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
