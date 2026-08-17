# Scenario

**Feature**: ahead merge-back with confirmation and --rm

```
user -> merge-back --rm --confirm-from-stdin Enter -> merge + remove
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
	req.Remove = true
	if req.WorktreePath == "" {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```