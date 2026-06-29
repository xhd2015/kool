# Scenario

**Feature**: extra positional arguments after directory

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DirPath = initValidDir(t, req.WorkingDir, "proj")
	req.ExtraPositional = []string{"extra"}
	return nil
}
```