# Scenario

**Feature**: list --json emits machine-readable JSON without ANSI

```
multi-task fixture + --json
  -> stdout JSON array/object of tasks; exit 0
```

## Steps

1. Multi-task fixture; JSON=true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.JSON = true
	return nil
}
```
