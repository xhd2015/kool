# Scenario

**Feature**: sessions restore dry-run plans Space placement from checkpoint

```
Caller
  -> seed ckpt with optional space / iterm_window_id
  -> sessions restore --dry-run [--ignore-macos-space]
  <- space N (Desktop N+1) lines or omit when ignore; not stamped
```

## Steps

1. ModeRestore; DryRun (shared by restore space leaves).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	req.DryRun = true
	return nil
}
```
