# Scenario

**Feature**: dry-run merge-back from detached HEAD ahead of main

```
user (detached HEAD, ahead) -> merge-back --dry-run -> planned merge commands, no mutations
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	req.Remove = false
	return nil
}
```