# Scenario

**Feature**: trailing slash stripped from normalized path in URI

```
# trailing slash removed before URI encoding
ValidateDirPath(dir/) -> BuildOpenURI
```

## Steps
1. Create directory.
2. Pass path with trailing slash.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "slash-dir")
	req.DirPath = dir + string(filepath.Separator)
	return nil
}
```