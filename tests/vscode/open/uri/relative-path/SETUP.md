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
)

func Setup(t *testing.T, req *Request) error {
	markUriTree()
	markRootTree()
	initValidDir(t, req.WorkingDir, "subdir")
	req.DirPath = "subdir"
	return nil
}
```