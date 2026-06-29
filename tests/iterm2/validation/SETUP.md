# Scenario

**Feature**: CLI rejects invalid arguments before osascript

```
kool iterm2 (bad argv) -> validation error -> stderr, exit 1
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "cli"
	return nil
}
```