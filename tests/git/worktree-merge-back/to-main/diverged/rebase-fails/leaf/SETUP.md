# Scenario

**Feature**: confirm rebase on conflicting diverged branches

```
user -> merge-back --confirm-from-stdin + Enter -> rebase conflict -> error
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
	req.Remove = false
	return nil
}
```