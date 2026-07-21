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
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRootTree()
	markSetTitleTree()
	req.InSession = true
	req.Window = true
	req.TitleSet = true
	return nil
}

// markSetTitleWindowTree keeps hierarchical child packages importing this package live.
func markSetTitleWindowTree() {}
```
