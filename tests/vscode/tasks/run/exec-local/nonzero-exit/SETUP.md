# Scenario

**Feature**: local backend surfaces non-zero child exit

```
run "Fail Fast" --backend=local
  -> command false -> exit ≠ 0
```

## Steps

1. failLeafJSONC; Query=`Fail Fast`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeTasksJSON(t, req.WorkingDir, failLeafJSONC)
	req.Dir = req.WorkingDir
	req.Query = "Fail Fast"
	return nil
}
```
