# Scenario

**Bug**: `v0.2.0` must increment to `v0.2.1` (primary production repro)

```
IncrementTag("v0.2.0") -> "v0.2.1"
```

## Steps

1. Set `req.Tag` to `v0.2.0`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Tag = "v0.2.0"
	return nil
}
```