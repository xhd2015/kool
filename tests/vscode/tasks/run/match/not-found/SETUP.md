# Scenario

**Feature**: run unknown label errors

```
run "zzzz-missing" --dry-run -> Error not found
```

## Steps

1. Multi-task; Query missing.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "zzzz-missing"
	return nil
}
```
