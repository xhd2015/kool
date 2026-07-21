# Scenario

**Feature**: --allow-failure exits on the first child failure

```
kool for-every --allow-failure --max-runs 5 10ms sh -c 'echo fail-once; exit 1'
  -> one fail-once line; non-zero; does not continue to max-runs
```

## Steps

1. Always-fail child; allow-failure; high max-runs safety net.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markLoopFailurePolicyTree()
	markLoopTree()
	markRootTree()
	req.AllowFailure = true
	req.MaxFailure = nil
	req.MaxRuns = intPtr(5) // safety; policy must stop at first failure
	req.Command = "sh"
	req.Args = []string{"-c", "echo fail-once; exit 1"}
	return nil
}
```
