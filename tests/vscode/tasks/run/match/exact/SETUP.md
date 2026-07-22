# Scenario

**Feature**: run exact label match with --dry-run

```
run Compile --dry-run -> plan for Compile exit 0
```

## Steps

1. Multi-task; Query exact `Compile`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "Compile"
	return nil
}
```
