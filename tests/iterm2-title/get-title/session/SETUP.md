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
import "testing"

func Setup(t *testing.T, req *Request) error {
	markGetTitleTree()
	markRootTree()
	req.InSession = true
	req.Window = false
	return nil
}

// markGetTitleSessionTree keeps hierarchical child packages importing this package live.
func markGetTitleSessionTree() {}
```
