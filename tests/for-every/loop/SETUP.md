# Scenario

**Feature**: for-every enters the run/sleep loop with explicit stop bounds

```
# always bounded for tests
user -> kool for-every[-<dur>] --max-runs N | --max-failure N | --allow-failure …
  -> iterate: run cmd; maybe stop; else sleep Interval; repeat
```

## Steps

1. Default short interval and at least one stop flag at leaf level.
2. Help remains false.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Help = false
	if req.Duration == "" {
		req.Duration = "10ms"
	}
	// Loop tests should finish quickly; keep process timeout modest.
	if req.ProcessTimeout <= 0 || req.ProcessTimeout > 10*time.Second {
		req.ProcessTimeout = 10 * time.Second
	}
	return nil
}

// markLoopTree keeps hierarchical child packages importing this package live.
func markLoopTree() {}
```
