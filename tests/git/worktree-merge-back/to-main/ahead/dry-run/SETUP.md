# Scenario

**Feature**: dry-run for ahead branch

```
user -> merge-back --dry-run -> planned commands only
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	req.Remove = false
	req.ConfirmFromStdin = false
	if req.WorktreePath == "" {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```