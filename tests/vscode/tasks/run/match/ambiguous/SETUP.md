# Scenario

**Feature**: run ambiguous substring errors listing matches

```
run alpha --dry-run with Alpha One + Alpha Two -> Error
```

## Steps

1. Ambiguous pair; Query=`alpha`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeTasksJSON(t, req.WorkingDir, ambiguousPairJSONC)
	req.Dir = req.WorkingDir
	req.Query = "alpha"
	return nil
}
```
