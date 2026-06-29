# Scenario

**Feature**: single --send follow-up

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DirPath = initValidDir(t, req.WorkingDir, "send-one")
	req.Send = []string{"grok"}
	return nil
}
```