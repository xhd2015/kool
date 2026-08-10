# Scenario

**Feature**: sealed primary with --runtime-load-devbox binds events sock while guest runs

```
kool sandbox build -o load.bin --file load.txt=load.txt
kool sandbox build -o sandbox.bin --file primary.txt=primary.txt \
  --runtime-load-devbox ABS_LOAD
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'echo ready > .guest-ready; …'
  -> PARENT/events/<session-id>.sock exists
  -> events dir present; sock mode preferably 0600
```

## Steps

1. Secondary pack with load.txt; primary RuntimeLoadDevbox=abs secondary.
2. Long guest (default ready/stop loop).
3. StopGuest + WaitGuestExit after sock poll (cleanup for isolation).

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
	secOut := "load-bind.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"load.txt=load.txt"},
	}}
	req.ExtraFiles = []string{"primary.txt=primary.txt"}
	req.RuntimeLoadDevbox = []string{secAbs}
	req.SealedLoadDevbox = nil
	req.NotifyAfterStart = false
	req.RebuildLoadAfterStart = false
	req.StopGuest = true
	req.WaitGuestExit = true
	return nil
}
```
