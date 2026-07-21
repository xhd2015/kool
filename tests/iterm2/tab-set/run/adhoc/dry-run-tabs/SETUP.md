# Scenario

**Feature**: ad-hoc --tab ×2 --dry-run prints plan without config file

```
run scratch --tab "echo alpha" --tab "echo beta" --dry-run
  -> exit 0; plan shows both commands; no <name>.json required
```

## Steps

1. Empty ConfigDir (no scratch.json).
2. SetName=scratch; Tabs two bare commands; DryRun=true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRunAdhocTree()
	markRunTree()
	markTabSetRunAdhocTree()
	markTabSetRunTree()
	markTabSetTree()
	markRootTree()
	req.SetName = "scratch"
	req.DryRun = true
	req.Tabs = []string{"echo alpha", "echo beta"}
	return nil
}
```
