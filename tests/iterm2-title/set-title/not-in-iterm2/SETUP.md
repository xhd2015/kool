# Scenario

**Feature**: set-title outside iTerm2 warns and fails without osascript

```
# no ITERM_SESSION_ID
kool iterm2 set-title my-title
  -> stderr "warning: nothing to set; needs to be run inside iTerm2"
  -> exit 1, no script
```

## Steps

1. Provide a non-empty title so the failure is session-env, not validation.
2. Clear in-session env (`InSession=false`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.InSession = false
	req.Title = "my-title"
	req.TitleSet = true
	return nil
}
```
