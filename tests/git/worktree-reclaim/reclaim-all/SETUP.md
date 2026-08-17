# Scenario

**Feature**: reclaim all linked worktrees of the current repository

```
# user passes --all; handler resolves main repo and iterates linked worktrees
user -> kool git worktree reclaim --all -> reclaim handler -> all linked candidates
```

## Context

- Skipped worktrees do not fail the command; only removal errors cause non-zero exit
- Cwd may be the main repo or any linked worktree directory

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.All = true
	req.Path = ""
	return nil
}
```