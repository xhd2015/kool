# Scenario

**Feature**: diverged rebase+merge with --rm removes worktree and branch

```
user -> merge-back --rm --confirm-from-stdin + Enter -> rebase + merge + remove
```

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