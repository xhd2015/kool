# Scenario

**Feature**: multi-arg child command is passed through intact

```
kool for-every --max-runs 1 10ms echo alpha beta gamma
  -> stdout: alpha beta gamma\n ; exit 0
```

## Steps

1. One run; echo with three args.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markLoopSpacedTree()
	markLoopTree()
	markRootTree()
	req.Duration = "10ms"
	req.MaxRuns = intPtr(1)
	req.Command = "echo"
	req.Args = []string{"alpha", "beta", "gamma"}
	return nil
}
```
