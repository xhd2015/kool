# Scenario

**Feature**: invalid refA fails before comparison

```
# nonexistent-branch cannot be resolved
compare_branch.Handle(refA=nonexistent-branch, refB=main) -> error
```

## Steps
- Set RefA to a non-existent reference `nonexistent-branch`
- Set RefB to `main` (a valid ref)

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.RefA = "nonexistent-branch"
	req.RefB = "main"
	return nil
}
```
