# Scenario

**Feature**: reclaim detached HEAD worktree whose commit is included in main

```
# detach HEAD at merged commit, then reclaim
reclaim handler -> git worktree remove
```

## Steps

1. Detach HEAD in the merged feature worktree inherited from clean-merged setup

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	runGit(t, req.WorktreePath, "checkout", "--detach", "HEAD")
	req.DryRun = false
	return nil
}
```