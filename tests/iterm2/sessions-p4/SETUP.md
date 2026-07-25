# Scenario

**Feature**: kool wires agent-pro procresolve for busy iTerm panes (P4)

```
# default snapshot enrich
sessions snapshot --no-color
  -> Capture (phased) + process idle/busy
  -> ResolveFromPID inject (AgentResolveByTTY) on busy tty
  -> CLI: grok/codex session id + FormatTree Unicode connectors

# flags
  --no-enrich  -> skip resolve attach
  --no-tree    -> session id only (no ├──/└──/│)
  --json       -> agent.session_id on session objects

# session status parity
session <busy-id> status --no-color
  -> same resolve + render for one session
```

## Preconditions

- Nested doctest root at `tests/iterm2/sessions-p4/` (module root =
  `filepath.Join(d.DOCTEST_ROOT, "..", "..")`).
- P3 sessions tree (`./tests/iterm2/sessions`) stays GREEN and is not modified.
- Package `github.com/xhd2015/kool/tools/iterm2` **P4 surface** (see root
  `DOCTEST.md` DSN):
  - `AgentResolveFixture`, `AgentTreeNode`
  - `PhasedFixtureOpts.AgentResolveByTTY`
  - CLI `--no-enrich`, `--no-tree`
  - `SnapshotSession.Agent` / JSON `agent.session_id`
- Injectable resolve only — **no live lsof**, no real agent-run/grok.
- agent-pro path for implementer: `../agent-pro-master-2026-07-25`
  (`github.com/xhd2015/agent-pro/pkgs/procresolve`).

## Steps

1. Root `Setup` creates an isolated `WorkingDir`.
2. Grouping/leaf `Setup` sets `Mode`, flags, and `AgentResolveByTTY`.
3. `Run` installs fixture collector + invokes CLI.

## Context

- Canonical hierarchy: Win-A idle `ttys001`, Win-B busy `ttys002`
  (same ids as P3).
- Agent fixture uuid: `019fabcdef-1234-5678-9abc-def012345678`.
- Fixed clock `2026-07-25T12:00:00Z`, host `testhost`.

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
	return nil
}
```
