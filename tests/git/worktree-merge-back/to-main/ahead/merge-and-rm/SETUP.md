# Scenario

**Feature**: ahead merge-back with confirmation and --rm

```
user -> merge-back --rm --confirm-from-stdin Enter -> merge + remove
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfirmFromStdin = true
	req.StdinInput = "\n"
	req.Remove = true
	if req.WorktreePath == "" {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```