# Scenario

**Feature**: merge ahead branch into sibling worktree HEAD

```
user -> merge-back --to sibling --confirm-from-stdin + Enter -> sibling HEAD advanced
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfirmFromStdin = true
	req.StdinInput = "\n"
	req.Remove = false
	req.DryRun = false
	return nil
}
```