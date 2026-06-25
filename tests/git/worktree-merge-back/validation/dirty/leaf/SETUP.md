# Scenario

**Feature**: run merge-back on dirty worktree

```
user (cwd=dirty wt) -> merge-back handler -> uncommitted changes error
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatal("expected dirty worktree from ancestor setup")
	}
	return nil
}
```