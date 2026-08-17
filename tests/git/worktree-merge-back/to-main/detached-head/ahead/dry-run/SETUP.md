# Scenario

**Feature**: dry-run merge-back from detached HEAD ahead of main

```
user (detached HEAD, ahead) -> merge-back --dry-run -> planned merge commands, no mutations
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
	return nil
}
```