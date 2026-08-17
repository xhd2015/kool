# Scenario

**Feature**: ahead branch without TTY confirmation is rejected

```
user (non-TTY) -> merge-back -> confirmation required error, no mutations
```

## Steps

1. Run merge-back without `--confirm-from-stdin` and without piped stdin

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ConfirmFromStdin = false
	req.StdinInput = ""
	req.Remove = false
	return nil
}
```