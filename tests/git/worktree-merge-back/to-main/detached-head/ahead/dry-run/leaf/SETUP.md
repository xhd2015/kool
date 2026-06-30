# Scenario

**Feature**: dry-run leaf for detached HEAD ahead of main

```
user -> merge-back --dry-run -> planned commands only
```

## Steps

1. Run merge-back with `--dry-run` only (configured by ancestor setup)

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	req.Remove = false
	req.ConfirmFromStdin = false
	if req.WorktreePath == "" {
		t.Fatal("expected detached worktree from ancestor setup")
	}
	return nil
}
```