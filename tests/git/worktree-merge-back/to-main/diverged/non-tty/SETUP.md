# Scenario

**Feature**: diverged merge-back without TTY confirmation

```
user (non-TTY) -> merge-back -> error before mutations
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfirmFromStdin = false
	req.StdinInput = ""
	req.Remove = false
	if req.WorktreePath == "" {
		t.Fatal("expected diverged worktree from ancestor setup")
	}
	return nil
}
```