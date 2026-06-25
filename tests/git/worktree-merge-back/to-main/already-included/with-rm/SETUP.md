# Scenario

**Feature**: already-included merge-back with --rm

```
user -> merge-back --rm -> remove worktree + delete branch
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Remove = true
	req.DryRun = false
	if req.WorktreePath == "" {
		t.Fatal("expected included worktree from ancestor setup")
	}
	return nil
}
```