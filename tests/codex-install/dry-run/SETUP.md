# Scenario

**Feature**: install dry-run plans action without shell mutation

```
# dry-run inspects presence/versions then prints plan
user -> kool codex install --dry-run
  -> plan install | update | noop on stdout; exit 0; ShellCalls empty
```

## Steps

1. Enable DryRun for all children.
2. Leaves set Present + LocalRaw + Latest.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.DryRun = true
	req.CheckUpdate = false
	req.Help = false
	return nil
}
```
