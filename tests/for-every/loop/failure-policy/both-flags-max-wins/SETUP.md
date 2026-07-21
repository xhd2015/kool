# Scenario

**Feature**: when both failure flags set, --max-failure wins over --allow-failure

```
kool for-every --allow-failure --max-failure 3 --max-runs 10 10ms sh -c 'echo fail-line; exit 1'
  -> three fail-line prints (not one); non-zero
```

## Steps

1. Both AllowFailure and MaxFailure=3; always-fail child; high max-runs.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markLoopFailurePolicyTree()
	markLoopTree()
	markRootTree()
	req.AllowFailure = true
	req.MaxFailure = intPtr(3)
	req.MaxRuns = intPtr(10)
	req.Command = "sh"
	req.Args = []string{"-c", "echo fail-line; exit 1"}
	return nil
}
```
