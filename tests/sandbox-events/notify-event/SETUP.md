# Scenario

**Feature**: `kool sandbox notify-event` publisher CLI (stateless Topology B)

```
user -> kool sandbox notify-event --type TYPE --path ABS [--root DIR] [--dry-run] [-h]
  -> list $ROOT/events/*.sock; dial / dry-run / help
```

## Steps

1. Default: `RunNotifyEvent=true` for non-help leaves under this group.
2. Resolve `--root` / `KOOL_SANDBOX_ROOT` under leaf `WorkingDir` fixtures.
3. Help leaf overrides with `HelpNotifyEvent=true`.

```go
import (
	"os"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Notify leaves finish quickly.
	if req.ProcessTimeout > 30*time.Second || req.ProcessTimeout <= 0 {
		req.ProcessTimeout = 30 * time.Second
	}
	req.LiveSession = false
	req.HelpNotifyEvent = false
	req.RunNotifyEvent = true
	// Short sandbox parent so AF_UNIX socks under parent/events fit macOS
	// sun_path (~104). t.TempDir() paths are often too long for mock listeners.
	if req.SandboxRootParent == "" {
		dir, err := os.MkdirTemp("/tmp", "ksb-")
		if err != nil {
			return err
		}
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		req.SandboxRootParent = dir
	}
	if req.EventRoot == "" {
		req.EventRoot = req.SandboxRootParent
		req.EventRootSet = true
	}
	if req.EventType == "" {
		req.EventType = "devbox.updated"
	}
	return nil
}
```
