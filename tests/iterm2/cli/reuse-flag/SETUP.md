# Scenario

**Feature**: `-r` reuses current session script path

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DirPath = initValidDir(t, req.WorkingDir, "reuse-target")
	req.Reuse = true
	return nil
}
```