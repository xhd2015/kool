# Scenario

**Feature**: CLI rejects invalid arguments before osascript

```
kool iterm2 (bad argv) -> validation error -> stderr, exit 1
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Phase = "cli"
	return nil
}

// markValidationTree keeps hierarchical child packages importing this package live.
func markValidationTree() {}
```