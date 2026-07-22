# Scenario

**Feature**: find with zero matches errors

```
query "zzzz-nope" -> Error exit ≠ 0
```

## Steps

1. Multi-task fixture; Query with no matches.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "zzzz-nope"
	return nil
}
```
