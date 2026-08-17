# Scenario

**Feature**: --list omits a `testdata/` submodule line (name skip)

```
# testdata is a reserved Go fixture dir; scan prunes it
root + testdata/x/go.mod -> kool go modules --list --dir <root> -> stdout:
  . some.com/root
  (no testdata line)
```

## Steps

1. Root workspace (`some.com/root`, git-init'd) is created by the `list/` grouping Setup.
2. Add `testdata/x/go.mod` (`some.com/root/testdata-x`) to `req.RootDir`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeModule(t, filepath.Join(req.RootDir, "testdata", "x"), "some.com/root/testdata-x")
	return nil
}
```
