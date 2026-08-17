# Scenario

**Feature**: diverged non-conflicting branches rebase and merge successfully

```
user -> merge-back --confirm-from-stdin + Enter -> rebase + ff merge into main
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