# Scenario

**Feature**: run merge-back --rm on dirty worktree

```
user (cwd=dirty wt) -> merge-back --rm -> uncommitted changes error
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatal("expected dirty worktree from ancestor setup")
	}
	return nil
}
```