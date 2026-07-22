# Scenario

**Feature**: show ambiguous CI substring errors listing matches

```
show "alpha" with Alpha One + Alpha Two -> Error ambiguous
```

## Steps

1. Ambiguous pair fixture; Query=`alpha` (not exact).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, ambiguousPairJSONC)
	req.Dir = req.WorkingDir
	req.Query = "alpha"
	return nil
}
```
