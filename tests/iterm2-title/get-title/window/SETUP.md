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
	req.InSession = true
	req.Window = true
	return nil
}
```
