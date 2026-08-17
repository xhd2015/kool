# Scenario

**Feature**: iTerm2 not installed (env override)

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Phase = "cli"
	req.DirPath = initValidDir(t, req.WorkingDir, "proj")
	req.InstalledEnv = "0"
	return nil
}
```