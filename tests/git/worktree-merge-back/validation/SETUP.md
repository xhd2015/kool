# Scenario

**Feature**: merge-back validation rejects invalid source or target before git mutations

```
# handler validates linked worktree, cleanliness, and target resolution
user -> kool git worktree merge-back -> merge-back handler -> validation error
```

## Context

- Validation failures return non-zero exit and perform no git mutations

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = false
	req.Remove = false
	req.ConfirmFromStdin = false
	req.StdinInput = ""
	return nil
}
```