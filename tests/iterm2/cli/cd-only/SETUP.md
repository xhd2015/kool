# Scenario

**Feature**: cd-only open without follow-ups

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markCliTree()
	markRootTree()
	req.DirPath = initValidDir(t, req.WorkingDir, "cd-only-target")
	return nil
}
```