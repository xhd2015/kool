# Scenario

**Feature**: default smart-open scan matches `path` or `user.koolTargetDir`

```
kool iterm2 <dir>/ -> scan finds session via path OR user.koolTargetDir -> reuse window (tab), not new window
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Reuse = false
	req.DirPath = initValidDir(t, req.WorkingDir, "smart-scan-target")
	return nil
}
```