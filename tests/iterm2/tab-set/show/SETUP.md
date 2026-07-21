# Scenario

**Feature**: tab-set show prints one config

```
tab-set show <name> -> window_name + tabs id/command
```

## Steps

1. Subcommand `show`; leaves set SetName and fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markTabSetTree()
	markRootTree()
	req.Subcommand = "show"
	return nil
}

// markTabSetShowTree keeps hierarchical child packages importing this package live.
func markTabSetShowTree() {}

// markShowTree keeps hierarchical child packages importing this package live.
func markShowTree() {}
```
