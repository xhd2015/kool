# Scenario

**Feature**: default install Ensure mutates via RunShell when needed

```
# ensure (no --dry-run / --check-update)
user -> kool codex install
  -> missing  → RunShell(InstallCmd) once
  -> outdated → RunShell(UpdateCmd) once
  -> current  → no shell
```

## Steps

1. Default mode: Help/DryRun/CheckUpdate false for children.
2. Leaves set Present + LocalRaw + Latest.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Help = false
	req.DryRun = false
	req.CheckUpdate = false
	return nil
}
```
