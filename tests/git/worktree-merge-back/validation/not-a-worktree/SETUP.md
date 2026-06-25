# Scenario

**Feature**: merge-back rejects cwd that is not a linked worktree

```
# main repo checkout is not a linked worktree path
user (cwd=main repo) -> merge-back handler -> not a linked worktree
```

## Steps

1. Initialize a main repository only (no linked worktree)
2. Run merge-back with cwd set to the main repo

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := initMainRepo(t)
	req.MainRepo = mainRepo
	req.Cwd = mainRepo
	return nil
}
```