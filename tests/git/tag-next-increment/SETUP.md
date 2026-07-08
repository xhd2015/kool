# Scenario

**Bug**: IncrementTag must accept version tags whose trailing patch segment is zero

```
# caller supplies tag string; IncrementTag returns next tag in-process
caller -> IncrementTag(tag) -> next tag string or error
```

## Preconditions

- Pure unit tests — no git repository, filesystem fixtures, or CLI subprocess required.

## Steps

1. Leaf `Setup` sets `req.Tag` to the scenario input.
2. Root `Run` calls `git_tag_next.IncrementTag(req.Tag)` in-process.
3. Leaf `Assert` checks `resp.NextTag` and `resp.Err`.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.Tag != "" {
		return fmt.Errorf("root setup: Tag must be unset before leaf Setup, got %q", req.Tag)
	}
	return nil
}
```