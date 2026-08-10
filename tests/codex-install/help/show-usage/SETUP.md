# Scenario

**Feature**: `--help` lists `--dry-run` and `--check-update`

```
kool codex install --help
  -> exit 0
  -> stdout contains --dry-run
  -> stdout contains --check-update
  -> stdout ends with newline
```

## Steps

1. Inherit Help=true from parent (restate for leaf clarity).

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
