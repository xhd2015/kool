# Scenario

**Feature**: leading [ that fails props parse is an error

```
run scratch --tab "[not-valid-props] echo hi" --dry-run
  -> Error exit ≠ 0 (props parse failure)
```

## Steps

1. --tab value starts with `[` but props are not valid key=value list.
2. DryRun optional (parse should fail before run).

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
	// Leading '[' but body is not key=value props → parse error per locked rule.
	req.Tabs = []string{"[not-valid-props] echo hi"}
	return nil
}
```
