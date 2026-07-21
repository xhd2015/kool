# Scenario

**Feature**: --force without --save is an error

```
run scratch --tab "echo x" --force --dry-run
  -> Error (--force only valid with --save)
```

## Steps

1. Tabs set; Force=true; Save=false; DryRun to avoid iTerm if force were ignored.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRunTree()
	markTabSetRunTree()
	markTabSetTree()
	markRootTree()
	req.SetName = "scratch"
	req.Force = true
	req.Save = false
	req.DryRun = true
	req.Tabs = []string{"echo x"}
	return nil
}
```
