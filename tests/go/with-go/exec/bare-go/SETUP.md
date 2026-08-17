# Scenario

**Feature**: bare `go` runs `$GOROOT/bin/go` with GOROOT and PATH set

```
# fake $GOROOT/bin/go writes GOROOT and first PATH entry to child.out
args=["go"] -> ExecGoroot -> child script GOROOT=$abs PATH0=$abs/bin
```

## Steps

1. Write a fake `$GOROOT/bin/go` script.
2. Set args to `["go"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	writeFakeGo(t, req.Goroot)
	req.Args = []string{"go"}
	return nil
}
```
