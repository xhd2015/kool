# Scenario

**Feature**: spaced form missing command after duration

```
kool for-every --max-runs 1 10ms
  -> non-zero; requires command
```

## Steps

1. Spaced form (Glued=false).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Glued = false
	return nil
}
```
