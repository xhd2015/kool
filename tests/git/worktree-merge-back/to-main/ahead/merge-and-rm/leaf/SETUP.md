# Scenario

**Feature**: user confirms ahead merge with --rm

```
user -> merge-back --rm --confirm-from-stdin + Enter -> ff merge + remove worktree + delete branch
```

## Steps

1. Run merge-back with `--rm` and confirmation

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfirmFromStdin = true
	req.StdinInput = "\n"
	req.Remove = true
	return nil
}
```