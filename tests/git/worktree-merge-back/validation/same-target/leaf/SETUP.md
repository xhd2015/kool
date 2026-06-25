# Scenario

**Feature**: invoke merge-back with --to equal to source worktree

```
user -> merge-back --to <same> -> validation error
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.To == "" || req.To != req.WorktreePath {
		t.Fatal("expected --to same as source worktree from ancestor setup")
	}
	return nil
}
```