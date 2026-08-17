# Scenario

**Feature**: reclaim rejects worktree whose HEAD is ahead of main

```
# branch not included in main HEAD
user -> kool git worktree reclaim <ahead-wt> -> error
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```