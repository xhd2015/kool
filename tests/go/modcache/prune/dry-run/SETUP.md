# Scenario

**Feature**: prune --dry-run prints a plan and leaves files on disk

```
prune --dry-run -> would remove; v1.0.0 still present
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	return nil
}
```
