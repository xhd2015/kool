# Scenario

**Feature**: Topology B event bus + `kool sandbox notify-event` + live-session
runtime-load file hot-reload (PHASE-1)

```
# publisher
user -> kool sandbox notify-event --type TYPE --path ABS [--root DIR] [--dry-run]
  -> dial $ROOT/events/*.sock; send JSON; summary / warning; exit 0

# live sealed session
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin [--load-devbox ABS]... -- <long guest>
  -> bind PARENT/events/<session-id>.sock; on devbox.updated re-apply load Files
```

## Preconditions

- Module root is `d.DOCTEST_ROOT/../..` (this tree lives at `tests/sandbox-events/`).
- `go` is on PATH; `Run` session-builds `kool` into
  `$TMPDIR/kool-sandbox-events-doctest-<d.DOCTEST_SESSION_ID>/kool` under a file lock.
- One-process mode: `Run` takes named `d *session.Doctest`; helpers use
  `d.DOCTEST_ROOT` / `d.DOCTEST_SESSION_ID` (no bare free identifiers).
- Product `notify-event` + runner event listener + file hot-reload are **not**
  implemented yet → leaves **RED** until implementer lands them.
- Per-leaf isolation: root Setup assigns `WorkingDir = t.TempDir()` for fixtures.
- `SandboxRootParent` / events parent defaults to a **short** dir under
  `/tmp/ksb-*` (not under long `t.TempDir()`) so AF_UNIX socks fit macOS
  sun_path (~104). Product `EventsDir` also maps long roots to `/tmp/kse/<hash>`.
- Parallel-safe: no `t.Setenv` / `t.Chdir`; `KOOL_SANDBOX_ROOT` only on child
  `cmd.Env`.
- Live-session leaves use host GOOS/GOARCH sealed binaries (no `--goos linux`).

## Steps

1. Root `Setup` creates an isolated `WorkingDir` and default process timeout.
2. Grouping/leaf `Setup` selects help / notify-event / live-session mode and
   writes fixtures under `WorkingDir`.
3. `Run` builds/reuses session `kool`, executes the mode, and for live session
   starts a long guest, polls for `events/*.sock`, optionally rebuilds a load
   pack + notifies, snapshots session files, then stops the guest.

## Context

- Shared session cache:
  `$TMPDIR/kool-sandbox-events-doctest-<DOCTEST_SESSION_ID>/`
- Events layout: `$KOOL_SANDBOX_ROOT/events/<session-id>.sock` when AF_UNIX-safe;
  product may map long roots to `/tmp/kse/<hash>/`. Harness uses short `/tmp/ksb-*`.
- Event JSON: `{"v":1,"type":"devbox.updated","path":"/abs","ts":"<RFC3339>"}`
- Helper `writeLocalFile` prepares `--file` sources under `WorkingDir`.

```go
import (
	"os"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	if err := os.MkdirAll(req.WorkingDir, 0755); err != nil {
		return err
	}
	if req.ProcessTimeout <= 0 {
		req.ProcessTimeout = 3 * time.Minute
	}
	return nil
}
```
