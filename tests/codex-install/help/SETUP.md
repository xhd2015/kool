# Scenario

**Feature**: install help documents `--dry-run` and `--check-update`

```
# user asks for install help
user -> kool codex install --help
  -> usage on stdout, exit 0 (no LookPath / no shell)
```

## Steps

1. Mark Help; no presence inject required for asserts (Run still injects safe fakes).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Help = true
	req.DryRun = false
	req.CheckUpdate = false
	return nil
}
```
