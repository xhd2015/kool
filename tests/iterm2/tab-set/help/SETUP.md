# Scenario

**Feature**: tab-set help subcommand

```
kool iterm2 tab-set --help -> usage mentioning list/run and config
```

## Steps

1. Leaves set Help or Subcommand for help path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markTabSetTree()
	markRootTree()
	// Grouping: help leaves set Help=true (or equivalent).
	req.Subcommand = ""
	return nil
}

// markTabSetHelpTree keeps hierarchical child packages importing this package live.
func markTabSetHelpTree() {}

// markHelpTree keeps hierarchical child packages importing this package live.
func markHelpTree() {}
```

