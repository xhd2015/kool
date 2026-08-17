# Scenario

**Feature**: user confirms ahead merge with default Enter

```
user -> merge-back --confirm-from-stdin + Enter -> ff merge, worktree remains
```

## Steps

1. Run merge-back with `--confirm-from-stdin` and empty line (default Y)

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ConfirmFromStdin = true
	req.StdinInput = "\n"
	req.Remove = false
	return nil
}
```