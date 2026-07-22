# Scenario

**Feature**: find with multiple matches exits 0 listing all

```
ambiguous pair fixture; query "alpha" -> Alpha One + Alpha Two exit 0
```

## Steps

1. Write ambiguous pair; Query=`alpha`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeTasksJSON(t, req.WorkingDir, ambiguousPairJSONC)
	req.Dir = req.WorkingDir
	req.Query = "alpha"
	return nil
}
```
