# Scenario

**Feature**: --max-failure 2 stops after two consecutive failures

```
kool for-every --max-failure 2 --max-runs 10 10ms sh -c 'echo fail-line; exit 1'
  -> two fail-line prints; non-zero; stops before max-runs
```

## Steps

1. Always-fail; max-failure 2; high max-runs safety.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markLoopFailurePolicyTree()
	markLoopTree()
	markRootTree()
	req.MaxFailure = intPtr(2)
	req.AllowFailure = false
	req.MaxRuns = intPtr(10)
	req.Command = "sh"
	req.Args = []string{"-c", "echo fail-line; exit 1"}
	return nil
}
```
