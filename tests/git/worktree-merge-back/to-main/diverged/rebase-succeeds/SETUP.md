# Scenario

**Feature**: diverged non-conflicting rebase and merge succeeds

```
user -> merge-back --confirm-from-stdin Enter -> rebase + ff merge
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
		t.Fatal("expected diverged worktree from ancestor setup")
	}
	return nil
}
```