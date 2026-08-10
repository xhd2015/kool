# Scenario

**Feature**: live sealed sandbox session binds Topology B event sock and may
hot-reload runtime-load file layers on `devbox.updated`

```
kool sandbox build -o sandbox.bin --file … --runtime-load-devbox ABS_LOAD
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c '…long guest…'
  -> PARENT/events/<session-id>.sock while guest runs
  -> notify-event may re-apply load Files into session root
  -> sock unlinked on session end
```

## Steps

1. LiveSession=true; host GOOS sealed build; After start poll events/*.sock.
2. Default guest waits for `.guest-stop` and writes `.guest-ready`.
3. Leaves set SecondaryPacks / RuntimeLoadDevbox / notify + reload options.

```go
import (
	"os"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HelpNotifyEvent = false
	req.RunNotifyEvent = false
	req.LiveSession = true
	req.SealedDoubleDash = true
	if req.ProcessTimeout <= 0 || req.ProcessTimeout < 2*time.Minute {
		req.ProcessTimeout = 2 * time.Minute
	}
	if req.Output == "" {
		req.Output = "sandbox.bin"
		req.OutputSet = true
	}
	// Short sandbox parent so parent/events/<session>.sock fits macOS AF_UNIX
	// sun_path (~104). Long t.TempDir() KOOL_SANDBOX_ROOT made bind fail and
	// the sealed guest never reached .guest-ready.
	if req.SandboxRootParent == "" {
		dir, err := os.MkdirTemp("/tmp", "ksb-")
		if err != nil {
			return err
		}
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		req.SandboxRootParent = dir
	}
	if req.SockWait <= 0 {
		req.SockWait = 12 * time.Second
	}
	if req.SettleAfterNotify <= 0 {
		req.SettleAfterNotify = 400 * time.Millisecond
	}
	return nil
}
```
