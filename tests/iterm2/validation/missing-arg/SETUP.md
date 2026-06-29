# Scenario

**Feature**: missing directory argument

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DirPath = ""
	return nil
}
```