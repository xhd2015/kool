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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Remove = true
	req.DryRun = false
	return nil
}
```