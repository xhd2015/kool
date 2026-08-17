# Scenario

**Feature**: absolute directory path produces correct vscode:// URI

```
# absolute input normalized and encoded
ValidateDirPath(absPath) -> BuildOpenURI
```

## Steps
1. Create directory at absolute path under temp dir.
2. Build URI from absolute path.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dir := initValidDir(t, req.WorkingDir, "abs-dir")
	req.DirPath = dir
	return nil
}
```