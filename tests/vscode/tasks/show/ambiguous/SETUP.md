# Scenario

**Feature**: show ambiguous CI substring errors listing matches

```
show "alpha" with Alpha One + Alpha Two -> Error ambiguous
```

## Steps

1. Ambiguous pair fixture; Query=`alpha` (not exact).

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
