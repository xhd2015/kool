# Scenario

**Feature**: already-included branch with --rm removes worktree and branch

```
user -> merge-back --rm -> worktree remove + branch -D (no merge prompt)
```

## Steps

1. Run merge-back with `--rm` from included worktree

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Remove = true
	req.DryRun = false
	return nil
}
```