# Scenario

**Feature**: --file PATH is used for the checkpoint write and mentioned in output

```
Caller
  -> sessions auto-backup --once --file <abs under DOCTEST_CASE>/custom-auto.json
  -> critical fixture
  <- Saved; file exists at custom path; stdout mentions path basename
```

## Steps

1. ModeAutoBackup; Once; UseCriticalFixture; FilePath absolute under d.DOCTEST_CASE.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = ModeAutoBackup
	req.Once = true
	req.UseCriticalFixture = true
	req.FilePath = filepath.Join(d.DOCTEST_CASE, "custom-auto.json")
	return nil
}
```
