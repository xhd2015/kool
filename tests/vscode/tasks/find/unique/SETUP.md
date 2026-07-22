# Scenario

**Feature**: find with unique case-insensitive match

```
query "compile" (CI) matches "Compile" only -> exit 0
```

## Steps

1. Multi-task fixture; Query=`compile`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "compile"
	return nil
}
```
