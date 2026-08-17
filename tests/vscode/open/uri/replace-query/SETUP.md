# Scenario

**Feature**: `--replace` adds `replace=true` to vscode:// URI

```
# BuildOpenURI with replace flag
ValidateDirPath -> BuildOpenURI(replace=true) -> ...&replace=true
```

## Steps
1. Create valid directory.
2. Build URI with `Replace=true`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dir := initValidDir(t, req.WorkingDir, "uri-replace-target")
	req.DirPath = dir
	req.Replace = true
	return nil
}
```