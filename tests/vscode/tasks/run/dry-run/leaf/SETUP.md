# Scenario

**Feature**: dry-run single leaf shell task prints command plan

```
run Compile --dry-run -> exit 0; plan mentions go build / Compile
```

## Steps

1. Multi-task fixture; Query=Compile; DryRun from parent.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	req.Query = "Compile"
	return nil
}
```
