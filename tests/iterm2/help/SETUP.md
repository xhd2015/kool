# Scenario

**Feature**: CLI help output

```
kool iterm2 --help -> usage on stdout -> exit 0
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Phase = "cli"
	return nil
}

// markHelpTree keeps hierarchical child packages importing this package live.
func markHelpTree() {}
```