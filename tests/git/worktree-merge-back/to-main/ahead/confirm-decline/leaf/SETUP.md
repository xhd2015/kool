# Scenario

**Feature**: user declines ahead merge confirmation

```
user -> merge-back --confirm-from-stdin + 'n' -> abort, no git mutations
```

## Steps

1. Run merge-back with `--confirm-from-stdin` and stdin `n`

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfirmFromStdin = true
	req.StdinInput = "n\n"
	req.Remove = false
	return nil
}
```