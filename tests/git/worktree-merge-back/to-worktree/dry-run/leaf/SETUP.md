# Scenario

**Feature**: dry-run with --to sibling prints planned commands only

```
user -> merge-back --to sibling --dry-run -> planned commands, no mutations
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.Remove = false
	req.ConfirmFromStdin = false
	return nil
}
```