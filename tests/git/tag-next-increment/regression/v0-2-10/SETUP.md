# Scenario

**Feature**: `v0.2.10` increments to `v0.2.11` (multi-digit patch)

```
IncrementTag("v0.2.10") -> "v0.2.11"
```

## Steps

1. Set `req.Tag` to `v0.2.10`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Tag = "v0.2.10"
	return nil
}
```