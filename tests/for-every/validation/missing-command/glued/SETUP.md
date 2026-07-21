# Scenario

**Feature**: glued form missing command after flags

```
kool for-every-10ms --max-runs 1
  -> non-zero; requires command
```

## Steps

1. Glued form with duration suffix only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markValidationMissingCommandTree()
	markRootTree()
	markValidationTree()
	req.Glued = true
	return nil
}
```
