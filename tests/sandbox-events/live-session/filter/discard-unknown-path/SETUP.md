# Scenario

**Feature**: notify path not in session loads → session files unchanged

```
# load seal has reload-me.txt = old-filter-content
# rebuild load with new content BUT notify --path /abs/other-not-loaded.bin
  -> session reload-me.txt remains old-filter-content
```

## Steps

1. Secondary load with old content; primary RuntimeLoadDevbox.
2. RebuildLoadAfterStart to new content (tempting reload) but NotifyLoadPath =
   a different abs path never in the session load set.
3. Snapshot before notify optional; assert final content still old.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if _, err := writeLocalFile(t, req.WorkingDir, "primary.txt", "primary-body\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "load-old.txt", "old-filter-content\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "load-new.txt", "new-filter-content\n"); err != nil {
		return err
	}
	secOut := "load-filter.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"load-old.txt=reload-me.txt"},
	}}
	req.ExtraFiles = []string{"primary.txt=primary.txt"}
	req.RuntimeLoadDevbox = []string{secAbs}
	req.RebuildLoadAfterStart = true
	req.RebuildLoadFiles = []string{"load-new.txt=reload-me.txt"}
	req.NotifyAfterStart = true
	req.NotifyEventType = "devbox.updated"
	// Different absolute path — not in session runtime-load set.
	req.NotifyLoadPath = filepath.Join(req.WorkingDir, "other-not-loaded.bin")
	req.ReadSessionRel = "reload-me.txt"
	req.SnapshotFileBeforeNotify = true
	req.StopGuest = true
	req.WaitGuestExit = true
	return nil
}
```
