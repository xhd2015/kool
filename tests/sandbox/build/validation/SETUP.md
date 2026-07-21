# Scenario

**Feature**: build validation errors before producing a sealed binary

```
# invalid argv / empty pack → fail fast
user -> kool sandbox build …
  -> stderr error, non-zero exit; no successful seal required
```

## Steps

1. Validation-only: short process timeout; no inspect/twice.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.AfterBuildInspect = false
	req.BuildTwice = false
	if req.ProcessTimeout > 30*time.Second || req.ProcessTimeout <= 0 {
		req.ProcessTimeout = 30 * time.Second
	}
	return nil
}
```
