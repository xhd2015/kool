# Scenario

**Feature**: run with `--backend=local` executes leaf commands in-process

```
run <label> --backend=local
  -> expand plan -> spawn shell/process leaves sequentially
  -> child stdout/stderr visible; exit from last/failed child
```

## Steps

1. Backend=`local`; DryRun=false.
2. Leaves install echo/false fixtures and Query.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Backend = "local"
	req.DryRun = false
	return nil
}
```
