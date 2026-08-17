# Scenario

**Feature**: default URI omits `replace` query parameter

```
# BuildOpenURI without replace flag
ValidateDirPath -> BuildOpenURI(replace=false) -> no replace= in query
```

## Steps
1. Create valid directory.
2. Build URI with `Replace=false`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dir := initValidDir(t, req.WorkingDir, "uri-default-target")
	req.DirPath = dir
	req.Replace = false
	return nil
}
```