# Scenario

**Feature**: get-title --window prints the window title

```
# mock osascript stdout = Project Window
kool iterm2 get-title --window
  -> stdout "Project Window\n"
```

## Steps

1. Mock osascript stdout `Project Window`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.OsascriptStdout = "Project Window"
	return nil
}
```
