# Scenario

**Feature**: after guest exits, events sock is unlinked

```
# start live session with runtime-load; observe sock
# stop guest / short sleep guest exits
  -> SockExistsAfterExit == false
```

## Steps

1. Secondary + RuntimeLoadDevbox primary (same as binds-sock).
2. Default guest; StopGuest + WaitGuestExit.
3. Assert sock present after start and absent after exit.

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
	if _, err := writeLocalFile(t, req.WorkingDir, "load.txt", "load-body\n"); err != nil {
		return err
	}
	secOut := "load-cleanup.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"load.txt=load.txt"},
	}}
	req.ExtraFiles = []string{"primary.txt=primary.txt"}
	req.RuntimeLoadDevbox = []string{secAbs}
	req.NotifyAfterStart = false
	req.RebuildLoadAfterStart = false
	req.StopGuest = true
	req.WaitGuestExit = true
	return nil
}
```
