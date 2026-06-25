# Scenario

**Feature**: user confirms ahead merge with default Enter

```
user -> merge-back --confirm-from-stdin Enter -> ff merge
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfirmFromStdin = true
	req.StdinInput = "\n"
	req.Remove = false
	if req.WorktreePath == "" {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```