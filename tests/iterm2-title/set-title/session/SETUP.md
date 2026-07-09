# Scenario

**Feature**: set-title default target is the session/tab name

```
# default target = session name
kool iterm2 set-title <title> + ITERM_SESSION_ID
  -> AppleScript sets session name (not window-only)
  -> stdout title changed line
```

## Steps

1. In-session; default target (`Window=false`); title provided by leaves.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.InSession = true
	req.Window = false
	req.TitleSet = true
	return nil
}
```
