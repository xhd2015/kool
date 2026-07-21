# Scenario

**Feature**: duplicate ad-hoc tab ids are rejected

```
run scratch --tab "[id=same] echo one" --tab "[id=same] echo two" --dry-run
  -> Error exit ≠ 0 (duplicate id)
```

## Steps

1. Two --tab values with the same explicit id.
2. DryRun=true.

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
	req.Tabs = []string{
		"[id=same] echo one",
		"[id=same] echo two",
	}
	return nil
}
```
