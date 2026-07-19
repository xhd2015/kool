# Scenario

**Feature**: bare --tab commands default to tab-1, tab-2 (order)

```
run scratch --tab "echo one" --tab "echo two" --dry-run
  -> plan ids tab-1, tab-2 (1-based --tab order)
```

## Steps

1. Two bare command tabs; DryRun; no props block.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SetName = "scratch"
	req.DryRun = true
	req.Tabs = []string{"echo one", "echo two"}
	return nil
}
```
