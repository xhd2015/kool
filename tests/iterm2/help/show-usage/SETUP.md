# Scenario

**Feature**: --help prints usage

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "cli"
	req.Help = true
	return nil
}
```