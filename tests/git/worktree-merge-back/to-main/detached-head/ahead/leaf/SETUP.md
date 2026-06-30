# Scenario

**Feature**: merge-back merges detached HEAD commit into main after confirmation

```
user (detached HEAD, ahead) -> merge-back --confirm-from-stdin + Enter -> ff merge into main
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