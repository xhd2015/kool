# Scenario

**Feature**: show unknown label errors

```
show "No Such Task" -> Error exit ≠ 0
```

## Steps

1. Multi-task fixture; Query missing.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "No Such Task"
	return nil
}
```
