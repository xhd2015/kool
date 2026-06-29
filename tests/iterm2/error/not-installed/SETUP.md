# Scenario

**Feature**: iTerm2 not installed (env override)

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "cli"
	req.DirPath = initValidDir(t, req.WorkingDir, "proj")
	req.InstalledEnv = "0"
	return nil
}
```