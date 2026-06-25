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
)

func Setup(t *testing.T, req *Request) error {
	runGit(t, req.WorktreePath, "checkout", "--detach", "HEAD")
	req.DryRun = false
	return nil
}
```