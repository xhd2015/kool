# Scenario

**Feature**: user declines ahead merge confirmation

```
user -> merge-back --confirm-from-stdin 'n' -> abort
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ConfirmFromStdin = true
	req.StdinInput = "n\n"
	req.Remove = false
	if req.WorktreePath == "" {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```