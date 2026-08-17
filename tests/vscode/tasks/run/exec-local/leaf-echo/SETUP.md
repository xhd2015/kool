# Scenario

**Feature**: local backend runs a single shell echo; marker on stdout

```
run "Say Hello" --backend=local
  -> exit 0; stdout contains KOOL_TASKS_P2_HELLO
```

## Steps

1. echoLeafJSONC; Query=`Say Hello`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeTasksJSON(t, req.WorkingDir, echoLeafJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Say Hello"
	return nil
}
```
