# Scenario

**Feature**: default smart-open sends cd to the new tab in matchingWindow

```
kool iterm2 <dir> -> match branch: create tab in matchingWindow -> cd in THAT tab (not frontmost window)
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Reuse = false
	req.DirPath = initValidDir(t, req.WorkingDir, "smart-match-target")
	return nil
}
```