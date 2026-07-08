# Scenario

**Feature**: IncrementTag preserves existing behavior for non-zero trailing patches

```
# trailing patch > 0 — must continue to increment the last numeric segment
IncrementTag("v0.0.87") -> "v0.0.88"
IncrementTag("v0.2.1")  -> "v0.2.2"
IncrementTag("v0.2.10") -> "v0.2.11"
```

Regression guard: fixing zero-patch increment must not break tags that already work.

## Steps

1. Leaf `Setup` sets `req.Tag` to a non-zero-patch version tag.

```go
import (
	"fmt"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.Tag != "" && strings.HasSuffix(req.Tag, ".0") {
		return fmt.Errorf("regression grouping: tag %q ends with .0, belongs under zero-patch/", req.Tag)
	}
	return nil
}
```