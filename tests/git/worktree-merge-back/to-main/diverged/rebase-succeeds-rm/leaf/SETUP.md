# Scenario

**Feature**: diverged rebase+merge with --rm removes worktree and branch

```
user -> merge-back --rm --confirm-from-stdin + Enter -> rebase + merge + remove
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ConfirmFromStdin = true
	req.StdinInput = "\n"
	req.Remove = true
	return nil
}
```