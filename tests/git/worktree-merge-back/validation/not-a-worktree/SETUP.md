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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMainRepo(t)
	req.MainRepo = mainRepo
	req.Cwd = mainRepo
	return nil
}
```