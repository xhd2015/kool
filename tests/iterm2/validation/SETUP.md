# Scenario

**Feature**: CLI rejects invalid arguments before osascript

```
kool iterm2 (bad argv) -> validation error -> stderr, exit 1
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Phase = "cli"
	return nil
}
```