# Scenario

**Feature**: invoke merge-back with --to equal to source worktree

```
user -> merge-back --to <same> -> validation error
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.To == "" || req.To != req.WorktreePath {
		t.Fatal("expected --to same as source worktree from ancestor setup")
	}
	return nil
}
```