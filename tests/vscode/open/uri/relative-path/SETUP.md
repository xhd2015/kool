# Scenario

**Feature**: relative directory path resolved in vscode:// URI

```
# relative input joined with cwd before encoding
ValidateDirPath(relative) -> BuildOpenURI
```

## Steps
1. Create `subdir` under working dir.
2. Pass relative path `subdir`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	initValidDir(t, req.WorkingDir, "subdir")
	req.DirPath = "subdir"
	return nil
}
```