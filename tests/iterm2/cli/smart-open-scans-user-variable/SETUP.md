# Scenario

**Feature**: default smart-open scan matches `path` or `user.koolTargetDir`

```
kool iterm2 <dir>/ -> scan finds session via path OR user.koolTargetDir -> reuse window (tab), not new window
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Reuse = false
	req.DirPath = initValidDir(t, req.WorkingDir, "smart-scan-target")
	return nil
}
```