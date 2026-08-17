# Scenario

**Feature**: -C pointing to missing directory fails

```
# directory does not exist
compare_branch.Handle(dir=nonexistent) -> error
```

## Steps
- Set req.Dir to a non-existent directory path
- Set RefA and RefB to arbitrary values

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Dir = filepath.Join(req.Dir, "nonexistent")
	req.RefA = "main"
	req.RefB = "main"
	return nil
}
```
