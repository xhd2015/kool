# Scenario

**Feature**: repeatable --send flags preserve order

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markCliTree()
	markRootTree()
	req.DirPath = initValidDir(t, req.WorkingDir, "send-multi")
	req.Send = []string{"grok", "codex"}
	return nil
}
```