# Scenario

**Feature**: --list emits root + sub-dir modules as space-separated `<dir> <path>` lines

```
# root go.mod + sub-dir/go.mod under one git repo
root + sub-dir/go.mod -> kool go modules --list --dir <root> -> stdout:
  . some.com/root
  sub-dir some.com/root/sub
```

## Steps

1. Root workspace (`some.com/root`, git-init'd) is created by the `list/` grouping Setup.
2. Add `sub-dir/go.mod` (`some.com/root/sub`) to `req.RootDir`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeModule(t, filepath.Join(req.RootDir, "sub-dir"), "some.com/root/sub")
	return nil
}
```
