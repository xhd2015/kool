# Scenario

**Feature**: diverged non-conflicting rebase and merge succeeds

```
user -> merge-back --confirm-from-stdin Enter -> rebase + ff merge
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ConfirmFromStdin = true
	req.StdinInput = "\n"
	req.Remove = false
	if req.WorktreePath == "" {
		t.Fatal("expected diverged worktree from ancestor setup")
	}
	return nil
}
```