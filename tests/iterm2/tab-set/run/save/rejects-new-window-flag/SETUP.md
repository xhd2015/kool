# Scenario

**Feature**: -n / --new-window with --save is an error

```
run scratch --tab "echo x" --save -n -> Error (save-only; window flags unused)
```

## Steps

1. Save + Tabs + NewWindow; no Force needed.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SetName = "scratch"
	req.Save = true
	req.NewWindow = true
	req.Tabs = []string{"echo x"}
	return nil
}
```
