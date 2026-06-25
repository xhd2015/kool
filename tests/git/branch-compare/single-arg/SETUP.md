# Scenario

**Feature**: single-arg compare-branch defaults omitted refB to current branch or HEAD

```
# user passes only refA; handler fills refB from git state
user -> kool git compare-branch <refA> -> compare_branch.Handle

# refB resolved via git branch --show-current, or HEAD when detached
compare_branch.Handle -> git branch --show-current -> refB default
```

## Context

- Single-arg mode passes only RefA to the CLI; RefB remains empty in Request
- refB defaults to the current branch name when checked out on a named branch
- refB defaults to `HEAD` when in detached HEAD state

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.RefB = ""
	return nil
}
```