# Scenario

**Feature**: kool iterm2 sessions snapshot — multi-phase AppleScript + streaming CLI (P3)

```
# help
user -> kool iterm2 sessions -h
  -> usage: snapshot, --json, --no-stream

# phased collection
SnapshotCollector.ListWindows()
  -> [W1 Win-A, W2 Win-B]
SnapshotCollector.ListTabsAndSessions(1)
  -> tabs/sessions for W1
CaptureSnapshot()
  -> ListWindows once + ListTabsAndSessions per window + ps enrich

# streaming CLI (default)
sessions snapshot --no-color
  -> emit W1 block after window 1 enriched
  -> emit W2 block after window 2 enriched
  -> footer summary

# buffered
sessions snapshot --no-stream | --json | --markdown | --html | -o FILE
  -> full collect then one render (no progressive W1 during ListTabs)
```

## Preconditions

- Nested doctest root at `tests/iterm2/sessions/` (module root =
  `d.DOCTEST_ROOT/../../..` only if nested deeper; here
  `filepath.Join(d.DOCTEST_ROOT, "..", "..")` is module root).
- Package `github.com/xhd2015/kool/tools/iterm2` provides `RunForTest`,
  `CaptureSnapshot`, `SetSnapshotCollectorForTest` today.
- **P3 implementer surface** (see root `DOCTEST.md` DSN):
  - `(*SnapshotCollector).ListWindows` / `ListTabsAndSessions`
  - `ActiveSnapshotCollectorForTest`
  - `PhasedFixtureOpts` + `InstallPhasedFixtureCollectorForTest`
  - CLI flag `--no-stream` and default progressive CLI rendering
- No live iTerm2, no agent/procresolve enrich.

## Steps

1. Root `Setup` creates an isolated `WorkingDir`.
2. Grouping/leaf `Setup` sets `Mode`, flags, and fixture/probe options.
3. `Run` installs fixture collector when needed and invokes API or CLI.

## Context

- Canonical fixture: two windows (`Win-A` idle on `ttys001`, `Win-B` busy on
  `ttys002`); fixed clock `2026-07-25T12:00:00Z`, host `testhost`.
- Stream probe: `OnListTabs` for window 2 records whether stdout already has
  `W1`.
- Format-conflict leaf does not install a fixture (no capture).

```go
import (
	"os"
	"testing"

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
	// Default: CLI snapshot path uses the two-window fixture when Mode is set
	// by descendants. Mode left empty until a grouping Setup assigns it.
	return nil
}
```
