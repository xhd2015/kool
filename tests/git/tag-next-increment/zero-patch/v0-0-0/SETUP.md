# Scenario

**Bug**: `v0.0.0` must increment to `v0.0.1` (all-zero patch)

```
IncrementTag("v0.0.0") -> "v0.0.1"
```

## Steps

1. Set `req.Tag` to `v0.0.0`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Tag = "v0.0.0"
	return nil
}
```