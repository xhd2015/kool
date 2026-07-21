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
import "testing"

func Setup(t *testing.T, req *Request) error {
	markGetTitleTree()
	markRootTree()
	req.InSession = false
	return nil
}
```
