# Scenario

**Feature**: already-included merge-back without --rm

```
user -> merge-back (no --rm) -> noop, worktree kept
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Remove = false
	req.DryRun = false
	if req.WorktreePath == "" {
		t.Fatal("expected included worktree from ancestor setup")
	}
	return nil
}
```