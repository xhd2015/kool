# Scenario

**Feature**: `kool vscode open --replace <dir>` succeeds

```
# CLI parses --replace and opens directory
kool vscode open --replace <dir> -> OpenDir(replace=true) -> ok
```

## Steps
1. Create valid directory under temp working dir.
2. Run `kool vscode open --replace <dir>` with fake `code` on PATH.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dir := initValidDir(t, req.WorkingDir, "replace-cli-target")
	req.DirPath = dir
	req.Replace = true
	installCLIMockIPC(t, req)
	return nil
}
```