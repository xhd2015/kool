# Scenario

**Feature**: user declines ahead merge confirmation

```
user -> merge-back --confirm-from-stdin 'n' -> abort
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfirmFromStdin = true
	req.StdinInput = "n\n"
	req.Remove = false
	if req.WorktreePath == "" {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```