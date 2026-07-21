# Scenario

**Feature**: for-every help mode (no loop)

```
# user asks for help
user -> kool for-every --help
  -> handler prints usage, exit 0 (no duration parse, no loop)
```

## Steps

1. Fix Help=true for descendants.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRootTree()
	req.Help = true
	return nil
}

// markHelpTree keeps hierarchical child packages importing this package live.
func markHelpTree() {}
```
