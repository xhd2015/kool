# Scenario

**Feature**: non-git directory cannot be compared

```
# -C or cwd is not a git repository
user -> kool git compare-branch main main -> compare_branch.Handle -> error
```

## Steps
- Create a temporary directory that is NOT a git repository
- Set req.Dir to this non-git directory

```go
import (
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dir, err := os.MkdirTemp("", "kool-branch-compare-nogit-*")
	if err != nil {
		return err
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	req.Dir = dir
	return nil
}
```
