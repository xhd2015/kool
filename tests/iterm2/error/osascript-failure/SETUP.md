# Scenario

**Feature**: osascript non-zero exit surfaces to CLI

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "cli"
	req.DirPath = initValidDir(t, req.WorkingDir, "proj")
	req.InstalledEnv = "1"
	req.OsascriptExit = 1
	return nil
}
```