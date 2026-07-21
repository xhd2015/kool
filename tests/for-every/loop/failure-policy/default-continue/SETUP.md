# Scenario

**Feature**: default policy continues after child failures until max-runs

```
kool for-every --max-runs 3 10ms sh -c 'echo fail-line; exit 1'
  -> three fail-line prints; non-zero exit (last child failure)
```

## Steps

1. Always-failing child that still prints a marker line; max-runs 3; no failure flags.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markLoopFailurePolicyTree()
	markLoopTree()
	markRootTree()
	req.MaxRuns = intPtr(3)
	req.AllowFailure = false
	req.MaxFailure = nil
	req.Command = "sh"
	req.Args = []string{"-c", "echo fail-line; exit 1"}
	return nil
}
```
