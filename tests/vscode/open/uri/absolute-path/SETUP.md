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
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "abs-dir")
	req.DirPath = dir
	return nil
}
```