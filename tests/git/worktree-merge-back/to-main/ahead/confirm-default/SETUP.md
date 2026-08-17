# Scenario

**Feature**: user confirms ahead merge with default Enter

```
user -> merge-back --confirm-from-stdin Enter -> ff merge
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
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```