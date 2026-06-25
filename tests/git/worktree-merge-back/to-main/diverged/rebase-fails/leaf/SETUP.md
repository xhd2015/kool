# Scenario

**Feature**: confirm rebase on conflicting diverged branches

```
user -> merge-back --confirm-from-stdin + Enter -> rebase conflict -> error
```

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