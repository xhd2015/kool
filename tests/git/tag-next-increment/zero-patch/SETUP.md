# Scenario

**Bug**: IncrementTag rejects tags ending in patch `.0`

```
# trailing numeric segment is zero — must still increment to .1
IncrementTag("v0.2.0") -> "v0.2.1"
IncrementTag("v0.0.0") -> "v0.0.1"
```

These are the primary bug cases: `GetLatestVersionTag` accepts semver patch zero, but
`IncrementTag` currently errors with `non numeric tag` or `invalid tag`.

## Steps

1. Leaf `Setup` sets `req.Tag` to a zero-patch version tag.

```go
import (
	"fmt"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.Tag != "" && !strings.HasSuffix(req.Tag, ".0") {
		return fmt.Errorf("zero-patch grouping: tag %q does not end with .0", req.Tag)
	}
	return nil
}
```