# Scenario

**Feature**: get-title default target is the session/tab name

```
# default get session name
kool iterm2 get-title + ITERM_SESSION_ID
  -> print session name + "\n"
```

## Steps

1. In-session; `Window=false`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InSession = true
	req.Window = false
	return nil
}
```
