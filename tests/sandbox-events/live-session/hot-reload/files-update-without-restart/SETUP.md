# Scenario

**Feature**: rewrite load seal on disk + notify-event updates session file content

```
# secondary load.bin packs reload-me.txt = "old-load-content"
# primary seals --runtime-load-devbox ABS_LOAD; long guest
# rebuild load.bin with reload-me.txt = "new-load-content"
# kool sandbox notify-event --type devbox.updated --path ABS_LOAD --root PARENT
  -> PARENT/<session>/reload-me.txt content becomes new-load-content
  -> guest still running (not restarted); stop after read
```

## Steps

1. Secondary with old content; primary RuntimeLoadDevbox.
2. RebuildLoadAfterStart with new local file mapped to same sandbox relpath.
3. NotifyAfterStart; ReadSessionRel=reload-me.txt; StopGuest.

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
	if _, err := writeLocalFile(t, req.WorkingDir, "load-old.txt", "old-load-content\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "load-new.txt", "new-load-content\n"); err != nil {
		return err
	}
	secOut := "load-hot.bin"
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
	// NotifyLoadPath empty → SecondaryPaths[0] after rebuild
	req.ReadSessionRel = "reload-me.txt"
	req.SnapshotFileBeforeNotify = false
	req.StopGuest = true
	req.WaitGuestExit = true
	return nil
}
```
