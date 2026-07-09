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
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Command = "get-title"
	return nil
}
```
