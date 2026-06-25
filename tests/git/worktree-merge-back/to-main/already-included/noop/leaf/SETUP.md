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
)

func Setup(t *testing.T, req *Request) error {
	req.Remove = false
	req.DryRun = false
	return nil
}
```