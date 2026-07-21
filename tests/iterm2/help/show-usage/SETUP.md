# Scenario

**Feature**: --help prints usage

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markHelpTree()
	markRootTree()
	req.Phase = "cli"
	req.Help = true
	return nil
}
```