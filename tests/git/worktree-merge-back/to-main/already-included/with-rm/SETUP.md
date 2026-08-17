# Scenario

**Feature**: already-included merge-back with --rm

```
user -> merge-back --rm -> remove worktree + delete branch
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Remove = true
	req.DryRun = false
	if req.WorktreePath == "" {
		t.Fatal("expected included worktree from ancestor setup")
	}
	return nil
}
```