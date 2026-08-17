# Scenario

**Feature**: single worktree reclaim by path

```
# user passes one linked worktree directory
user -> kool git worktree reclaim <worktree-dir> -> reclaim handler -> single candidate
```

## Context

- Single-path mode requires `<dir>` to be a linked worktree, not the main checkout
- Non-reclaimable targets return a non-zero exit code

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.All = false
	return nil
}
```