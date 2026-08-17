# Scenario

**Feature**: get-title outside iTerm2 warns and fails without osascript

```
# no ITERM_SESSION_ID
kool iterm2 get-title
  -> stderr "warning: nothing to get; needs to be run inside iTerm2"
  -> exit 1, no script
```

## Steps

1. Clear session env; no extra args.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InSession = false
	return nil
}
```
