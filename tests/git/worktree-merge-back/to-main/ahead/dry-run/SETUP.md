# Scenario

**Feature**: dry-run for ahead branch

```
user -> merge-back --dry-run -> planned commands only
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.Remove = false
	req.ConfirmFromStdin = false
	if req.WorktreePath == "" {
		t.Fatal("expected ahead worktree from ancestor setup")
	}
	return nil
}
```