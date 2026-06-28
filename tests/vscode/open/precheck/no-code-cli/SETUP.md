# Scenario

**Feature**: missing `code` CLI blocks open

```
# code not on PATH
EnsureCodeCLI -> error (mentions code / PATH)
```

## Steps
1. Create valid directory.
2. Run CLI with empty PATH so `code` is unavailable.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := initValidDir(t, req.WorkingDir, "target")
	req.DirPath = dir
	req.CodeInPath = false
	return nil
}
```