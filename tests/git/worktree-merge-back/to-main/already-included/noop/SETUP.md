# Scenario

**Feature**: already-included merge-back without --rm

```
user -> merge-back (no --rm) -> noop, worktree kept
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Remove = false
	req.DryRun = false
	if req.WorktreePath == "" {
		t.Fatal("expected included worktree from ancestor setup")
	}
	return nil
}
```