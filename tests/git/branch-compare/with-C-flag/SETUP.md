# Scenario

**Feature**: -C flag selects the git repository directory

```
# compare runs inside directory given by -C
user -> kool git compare-branch main main -C <dir> -> compare_branch.Handle
```

## Steps
- Prepare a temporary directory as the working context for testing the `-C` flag

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir, err := os.MkdirTemp("", "kool-branch-compare-C-*")
	if err != nil {
		return err
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	req.Dir = dir
	return nil
}
```
