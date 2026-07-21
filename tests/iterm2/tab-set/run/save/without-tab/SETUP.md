# Scenario

**Feature**: --save without any --tab is an error

```
run bots --save -> Error; no write
```

## Steps

1. Optional bots fixture present (should not matter).
2. Save=true; Tabs empty; no DryRun required.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRunSaveTree()
	markRunTree()
	markTabSetRunSaveTree()
	markTabSetRunTree()
	markTabSetTree()
	markRootTree()
	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.Save = true
	// Tabs intentionally empty — config mode cannot --save.
	return nil
}
```
