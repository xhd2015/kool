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
)

func Setup(t *testing.T, req *Request) error {
	req.ConfirmFromStdin = true
	req.StdinInput = "\n"
	req.Remove = false
	return nil
}
```