# Scenario

**Feature**: diverged non-conflicting branches rebase and merge successfully

```
user -> merge-back --confirm-from-stdin + Enter -> rebase + ff merge into main
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