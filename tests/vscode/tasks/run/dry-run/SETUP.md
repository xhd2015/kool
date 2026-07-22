# Scenario

**Feature**: run --dry-run expands plan without process spawn

```
run <label> --dry-run
  -> plan workspace + steps; no child process / iTerm
```

## Steps

1. DryRun=true for all descendants.
2. Leaves set Query and specialized fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	return nil
}
```
