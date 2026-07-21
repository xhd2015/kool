# Scenario

**Feature**: tab-set list reads config directory

```
KOOL_ITERM2_TAB_SET_DIR -> tab-set list -> set names on stdout
```

## Steps

1. Subcommand `list`.
2. Leaves prepare empty dir or fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markTabSetTree()
	markRootTree()
	req.Subcommand = "list"
	return nil
}

// markTabSetListTree keeps hierarchical child packages importing this package live.
func markTabSetListTree() {}

// markListTree keeps hierarchical child packages importing this package live.
func markListTree() {}
```
