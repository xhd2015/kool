# Scenario

**Feature**: diverged merge-back without TTY confirmation

```
user (non-TTY) -> merge-back -> error before mutations
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ConfirmFromStdin = false
	req.StdinInput = ""
	req.Remove = false
	if req.WorktreePath == "" {
		t.Fatal("expected diverged worktree from ancestor setup")
	}
	return nil
}
```