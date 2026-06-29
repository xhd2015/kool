# Scenario

**Feature**: non-darwin platform rejected

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "handler"
	req.GoOS = "linux"
	req.DirPath = initValidDir(t, req.WorkingDir, "proj")
	return nil
}
```