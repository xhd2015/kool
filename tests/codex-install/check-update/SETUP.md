# Scenario

**Feature**: install --check-update reports status only (exit 0; no shell)

```
# status-only path
user -> kool codex install --check-update
  -> status missing | update available | up to date
  -> exit 0 always on success path (even when outdated/missing)
  -> ShellCalls empty
```

## Steps

1. Enable CheckUpdate for all children.
2. Leaves set Present + LocalRaw + Latest.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.CheckUpdate = true
	req.DryRun = false
	req.Help = false
	return nil
}
```
