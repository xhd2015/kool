# Scenario

**Feature**: --file PATH is used for the checkpoint write and mentioned in output

```
Caller
  -> sessions auto-backup --once --file custom-auto.json
  -> critical fixture
  <- Saved under WorkingDir (t.TempDir); stdout mentions path basename
```

## Steps

1. ModeAutoBackup; Once; UseCriticalFixture; FilePath relative name only.
   Harness resolves under WorkingDir (defaults to t.TempDir) — never write into
   the source leaf (DOCTEST_CASE).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeAutoBackup
	req.Once = true
	req.UseCriticalFixture = true
	// Relative → resolveFile joins WorkingDir (temp), not the git source tree.
	req.FilePath = "custom-auto.json"
	return nil
}
```
