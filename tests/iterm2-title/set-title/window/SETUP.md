# Scenario

**Feature**: set-title --window targets the window title

```
# --window target
kool iterm2 set-title --window <title> + ITERM_SESSION_ID
  -> AppleScript sets name of the window that contains the session
```

## Steps

1. In-session with `Window=true`; title provided by leaves.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InSession = true
	req.Window = true
	req.TitleSet = true
	return nil
}
```
