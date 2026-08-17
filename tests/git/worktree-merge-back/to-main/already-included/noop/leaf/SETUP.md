# Scenario

**Feature**: already-included branch without --rm is a no-op

```
user -> merge-back (no --rm) -> noop success, worktree kept
```

## Steps

1. Run merge-back without `--rm` from included worktree

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Remove = false
	req.DryRun = false
	return nil
}
```