# Scenario

**Feature**: get-title --window reads the window title

```
# --window get
kool iterm2 get-title --window + ITERM_SESSION_ID
  -> print window name + "\n"
```

## Steps

1. In-session; `Window=true`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markGetTitleTree()
	markRootTree()
	req.InSession = true
	req.Window = true
	return nil
}

// markGetTitleWindowTree keeps hierarchical child packages importing this package live.
func markGetTitleWindowTree() {}
```
