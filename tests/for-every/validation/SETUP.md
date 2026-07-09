# Scenario

**Feature**: for-every validation errors before the loop starts

```
# invalid argv → no iterations, no infinite hang
user -> kool for-every … (bad duration | missing command | bad flags)
  -> stderr error, non-zero exit
```

## Steps

1. Mark this branch as validation-only: no Help; short process timeout still applies.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Help = false
	// Validation must fail fast; keep a tight wall clock.
	if req.ProcessTimeout > 5*time.Second || req.ProcessTimeout <= 0 {
		req.ProcessTimeout = 5 * time.Second
	}
	return nil
}
```
