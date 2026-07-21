# Scenario

**Feature**: tab-set run with dry-run and flag validation

```
tab-set run <name> [--dry-run] [-n] [--no-new-window]
  -> dry-run plan | flag conflict error
```

## Steps

1. Subcommand `run`.
2. Leaves set DryRun / NewWindow / NoNewWindow and fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markTabSetTree()
	markRootTree()
	req.Subcommand = "run"
	return nil
}

// markTabSetRunTree keeps hierarchical child packages importing this package live.
func markTabSetRunTree() {}

// markRunTree keeps hierarchical child packages importing this package live.
func markRunTree() {}
```
