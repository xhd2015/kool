# Scenario

**Feature**: kool iterm2 help documents title subcommands

```
# help path
kool iterm2 --help
  -> usage on stdout lists set-title and get-title
```

## Steps

1. Enable Help flag for descendants.

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
