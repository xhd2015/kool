# Scenario

**Feature**: get-title prints the current session title

```
# mock osascript stdout = Tab Alpha
kool iterm2 get-title
  -> stdout "Tab Alpha\n"
  -> script targets session name for UUID
```

## Steps

1. Mock osascript stdout `Tab Alpha` (no trailing newline in mock; CLI adds `\n`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markGetTitleSessionTree()
	markGetTitleTree()
	markRootTree()
	req.OsascriptStdout = "Tab Alpha"
	return nil
}
```
