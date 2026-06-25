# Scenario

**Feature**: kool git compare-branch compares two git references

```
# user invokes compare-branch; handler parses refs and optional -C
user -> kool git compare-branch <refA> [<refB>] [-C dir] -> compare_branch.Handle

# git resolves refs and reports relationship on stdout
compare_branch.Handle -> git.CompareBranches -> stdout report
```

## Preconditions

- The `kool` command is available in PATH
- Git is available in PATH

## Steps

1. Verify kool is available
2. Execute `kool git compare-branch` with RefA, optional RefB, and optional `-C <Dir>`
3. Capture stdout, stderr, and exit code

## Context

- `kool git compare-branch` compares two git references and reports their relationship
- Three outcomes: identical (same commit), fast-forward (one is ancestor of the other), or divergent (both have unique commits)
- When RefB is empty in Request, the CLI receives only RefA (single-arg mode)
- Supports `-C <dir>` flag to specify the git repository directory

```go
import (
	"fmt"
	"os/exec"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	_, err := exec.LookPath("kool")
	if err != nil {
		return fmt.Errorf("kool not found in PATH, build it first: %w", err)
	}
	return nil
}
```