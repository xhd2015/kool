# Scenario

**Feature**: --max-runs must be > 0 when provided

```
kool for-every --max-runs 0 10ms true
  -> non-zero validation; no loop
```

## Steps

1. Spaced form with valid duration and command; MaxRuns=0.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Glued = false
	req.Duration = "10ms"
	req.MaxRuns = intPtr(0)
	req.Command = "true"
	return nil
}
```
