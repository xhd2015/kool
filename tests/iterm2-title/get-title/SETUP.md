# Scenario

**Feature**: get-title subcommand prints session or window title

```
# get-title pipeline
kool iterm2 get-title [--window] + ITERM_SESSION_ID
  -> osascript returns current title -> stdout title + "\n"
```

## Steps

1. Fix `Command` to `get-title` for all descendants.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Command = "get-title"
	return nil
}
```
