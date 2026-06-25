# Scenario

**Feature**: invalid refB fails before comparison

```
# nonexistent-branch cannot be resolved
compare_branch.Handle(refA=main, refB=nonexistent-branch) -> error
```

## Steps
- Set RefA to `main` (a valid ref)
- Set RefB to a non-existent reference `nonexistent-branch`

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.RefA = "main"
	req.RefB = "nonexistent-branch"
	return nil
}
```
