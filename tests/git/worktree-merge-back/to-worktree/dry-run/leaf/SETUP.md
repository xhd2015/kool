# Scenario

**Feature**: dry-run with --to sibling prints planned commands only

```
user -> merge-back --to sibling --dry-run -> planned commands, no mutations
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	req.Remove = false
	req.ConfirmFromStdin = false
	return nil
}
```