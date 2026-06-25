# Scenario

**Feature**: reclaim rejects worktree whose HEAD is ahead of main

```
# branch not included in main HEAD
user -> kool git worktree reclaim <ahead-wt> -> error
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorktreePath == "" || !pathExists(t, req.WorktreePath) {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```