# kool iterm2 sessions — procresolve agent enrich + tree (P4)

**P4** wires agent-pro `pkgs/procresolve` into kool `tools/iterm2` so busy
panes show **grok/codex session ids** and a Unicode **process tree**
(`├──` / `└──` / `│` via `procresolve.FormatTree`). Injectable resolve for
tests — **no live `lsof`**.

**Sibling of** `./tests/iterm2/sessions/` (P3 phased/stream — **GREEN**,
untouched). This nested root is self-contained Classic TDD for P4 only.

**Out of scope:** soft cwd heuristics; cmdline `--session-id` as primary id;
P5 docs polish beyond help flags.

## Version

0.0.2

**Classic TDD (P4):** leaves are intentionally **RED** until the implementer
lands agent-pro require/replace, injectable resolve on the collector, CLI flags
`--no-enrich` / `--no-tree`, session model + render fields, and
`session status` parity.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — `kool iterm2 sessions snapshot …` or
  `kool iterm2 session <id> status …`.
- **kool CLI** — `tools/iterm2` snapshot + session status; format flags from
  P3 plus **`--no-enrich`** and **`--no-tree`**.
- **SnapshotCollector** — P3 phased hierarchy + process idle/busy; **P4**
  calls injectable resolve after process enrich on interesting root pid(s)
  (chosen/fg and/or agent-run roots).
- **procresolve (agent-pro)** —
  `ResolveFromPID(pid, Options) (*Result, error)` and
  `FormatTree(nodes, opts)` (Unicode connectors by default).
- **Fixture installer (tests)** —
  `InstallPhasedFixtureCollectorForTest` + **agent resolve map by tty**
  (no live processes / no real lsof).

### Behaviors

**Commands**

```text
kool iterm2 sessions -h|--help
kool iterm2 sessions snapshot [options]
kool iterm2 session <session-id> status [options]
```

**P4 flags (snapshot + status)**

| Flag | Effect |
|------|--------|
| (default) | On busy sessions with a resolvable pid, run procresolve; show agent session line(s) + tree |
| `--no-enrich` | Skip procresolve entirely (no agent session id / tree lines) |
| `--no-tree` | Keep agent session id (and optional title); omit FormatTree lines |
| `--json` | Include agent object on each enriched session (see model) |
| `--no-color` | No ANSI (tests always set this for CLI leaves) |

**Expected implementer package surface (`tools/iterm2`)**

```go
// Fixture types for injectable resolve (no live lsof).
type AgentTreeNode struct {
    PID  int    `json:"pid"`
    PPID int    `json:"ppid"`
    Role string `json:"role,omitempty"` // input | agent-run | … | grok | codex | other
    Cmd  string `json:"cmd"`
}

type AgentResolveFixture struct {
    Kind      string          // grok | codex | none
    SessionID string
    Title     string          // optional GrokTitle
    Tree      []AgentTreeNode // rendered via procresolve.FormatTree (Unicode)
}

// PhasedFixtureOpts gains (in addition to P3 fields):
//   AgentResolveByTTY map[string]AgentResolveFixture
//     key = short tty e.g. "ttys002"
//     applied after process enrich for sessions on that tty when enrich is on
//
// Production Capture path (conceptual):
//   if !NoEnrich && session has interesting pid:
//     res, err := procresolve.ResolveFromPID(pid, opts) // opts.ListProcs/Lsof injectable
//     attach Agent on SnapshotSession; CLI uses FormatTree(res.Tree)
//
// SnapshotSession gains:
//   Agent *SessionAgent `json:"agent,omitempty"`
//
// type SessionAgent struct {
//   Kind      string          `json:"kind"`
//   SessionID string          `json:"session_id"`
//   Title     string          `json:"title,omitempty"`
//   Tree      []AgentTreeNode `json:"tree,omitempty"`
// }
//
// CLI flags on sessions snapshot + session status: --no-enrich, --no-tree
// go.mod: require + replace github.com/xhd2015/agent-pro → ../agent-pro-master-2026-07-25
```

**CLI agent block (observable contract)**

For an enriched busy pane, human CLI includes:

1. Existing session line (status, iTerm id, pid, command) — P3
2. A line containing the **runner kind** (`grok` or `codex`) and the
   **agent session id** (fixture uuid)
3. Optional title line when `Title` / GrokTitle is non-empty
4. Unless `--no-tree`: process tree lines with Unicode connectors
   `├──`, `└──`, and `│` when the tree has siblings / depth (via FormatTree)

`--no-enrich` must not emit the fixture agent session id.

**JSON contract**

Under each enriched session object:

```json
"agent": {
  "kind": "grok",
  "session_id": "019fabcdef-1234-5678-9abc-def012345678",
  "title": "…",
  "tree": [ { "pid": 200, "ppid": 1, "role": "input", "cmd": "…" }, … ]
}
```

Field name `agent.session_id` is locked for asserts (equivalent nesting allowed
only if still addressable as session → agent → session_id).

**Inject / no live lsof**

Tests set `PhasedFixtureOpts.AgentResolveByTTY` only. Never requires real
agent-run, grok, or `lsof`.

### Locked fixture ids

| Name | Value |
|------|--------|
| Grok agent session | `019fabcdef-1234-5678-9abc-def012345678` |
| Grok title (optional) | `fixture-grok-title` |
| iTerm busy session UUID | `BBBBBBBB-0000-0000-0000-000000000002` |
| iTerm idle session UUID | `AAAAAAAA-0000-0000-0000-000000000001` |

## Decision Tree

```
sessions-p4/                            [nested DOCTEST root — P4]
├── help/
│   ├── enrich-flags/                   --no-enrich + tree flag on sessions -h
│   └── root-mentions-sessions/         iterm2 -h indexes snapshot + session status (P5)
├── enrich/                             [snapshot + injectable resolve]
│   ├── busy-grok/                      default enrich: session id + connectors
│   ├── agent-run-tree/                 3-level tree ├──/└──/│ + session id
│   ├── no-enrich/                      --no-enrich → no agent session id
│   ├── no-tree/                        --no-tree → id yes, connectors no
│   └── json/                           --json agent.session_id (+ tree)
└── status/
    └── with-resolve/                   session status shows runner session
```

Parameter ranking (most → least significant):

1. **Command surface** — help / snapshot enrich / session status
2. **Enrich on/off** — default vs `--no-enrich`
3. **Resolve topology** — simple busy grok vs agent-run multi-level tree
4. **Tree on/off** — default tree vs `--no-tree`
5. **Format** — CLI vs JSON

## Test Index

| Leaf | Description | Expect |
|------|-------------|--------|
| `help/enrich-flags/` | Help mentions `--no-enrich` and tree-related flag | RED |
| `help/root-mentions-sessions/` | Root `iterm2 -h` mentions sessions, snapshot, session status; trailing newline | RED (P5: content present, missing trailing newline) |
| `enrich/busy-grok/` | Busy fixture + resolve → CLI has session id + tree connectors | RED |
| `enrich/agent-run-tree/` | Agent-run tree fixture → `├──`/`└──`/`│` + session id | RED |
| `enrich/no-enrich/` | `--no-enrich` suppresses agent session id | RED |
| `enrich/no-tree/` | `--no-tree` keeps session id, omits connectors | RED |
| `enrich/json/` | `--json` has `agent.session_id` (and tree when present) | RED |
| `status/with-resolve/` | `session <busy-id> status` shows runner session id | RED |

## How to Run

```sh
doctest vet ./tests/iterm2/sessions-p4
doctest test ./tests/iterm2/sessions-p4

# P3 must remain green:
doctest test ./tests/iterm2/sessions
```

Nested root: no inheritance from `./tests/iterm2/sessions` P3 `Request`/`Run`.
Designer does **not** implement production enrich code.

```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

// Mode selects which surface Run exercises.
const (
	ModeHelp          = "help"
	ModeSnapshotCLI   = "snapshot-cli"
	ModeSessionStatus = "session-status"
)

// Locked fixture constants (procresolve + iTerm).
const (
	fixtureGrokSessionID = "019fabcdef-1234-5678-9abc-def012345678"
	fixtureGrokTitle     = "fixture-grok-title"
	fixtureITermBusyID   = "BBBBBBBB-0000-0000-0000-000000000002"
	fixtureITermIdleID   = "AAAAAAAA-0000-0000-0000-000000000001"
)

// Request drives one in-process snapshot / status / help scenario.
type Request struct {
	// Mode is help | snapshot-cli | session-status.
	Mode string

	// HelpArgs when ModeHelp (default: sessions -h).
	HelpArgs []string

	// Snapshot / status CLI flags.
	JSON     bool
	NoColor  bool
	NoStream bool
	NoEnrich bool
	NoTree   bool
	// ExtraArgs appended after known flags (escape hatch).
	ExtraArgs []string

	// SessionRef for ModeSessionStatus (iTerm unique id / prefix / tty).
	SessionRef string

	// AgentResolveByTTY installs injectable resolve fixtures (P4).
	// Keys are short tty names (e.g. "ttys002"). Empty → no agent enrich inject.
	AgentResolveByTTY map[string]iterm2.AgentResolveFixture

	// UseTwoWindowFixture installs the canonical two-window hierarchy (P3 shape).
	UseTwoWindowFixture bool

	// ITermRunning nil → true; false → not-running (unused by P4 leaves).
	ITermRunning *bool

	WorkingDir string
}

// Response is CLI / help output after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func boolPtr(v bool) *bool { return &v }

// fixtureWindows is the canonical two-window hierarchy (same shape as P3).
func fixtureWindows() []iterm2.SnapshotWindow {
	return []iterm2.SnapshotWindow{
		{
			Index: 1,
			Name:  "Win-A",
			Tabs: []iterm2.SnapshotTab{
				{
					Index: 1,
					Name:  "Tab-A",
					Sessions: []iterm2.SnapshotSession{
						{
							Index:   1,
							ID:      fixtureITermIdleID,
							Name:    "idle-sess",
							TTY:     "/dev/ttys001",
							Profile: "Default",
						},
					},
				},
			},
		},
		{
			Index: 2,
			Name:  "Win-B",
			Tabs: []iterm2.SnapshotTab{
				{
					Index: 1,
					Name:  "Tab-B",
					Sessions: []iterm2.SnapshotSession{
						{
							Index:   1,
							ID:      fixtureITermBusyID,
							Name:    "busy-sess",
							TTY:     "/dev/ttys002",
							Profile: "Default",
						},
					},
				},
			},
		},
	}
}

// busyGrokResolve: shell → grok (Unicode last-child connector └──).
func busyGrokResolve() iterm2.AgentResolveFixture {
	return iterm2.AgentResolveFixture{
		Kind:      "grok",
		SessionID: fixtureGrokSessionID,
		Title:     fixtureGrokTitle,
		Tree: []iterm2.AgentTreeNode{
			{PID: 100, PPID: 1, Role: "input", Cmd: "-zsh"},
			{PID: 101, PPID: 100, Role: "grok", Cmd: "/usr/local/bin/grok"},
		},
	}
}

// agentRunTreeResolve: agent-run → serve → grok (├── / │ / └──).
func agentRunTreeResolve() iterm2.AgentResolveFixture {
	return iterm2.AgentResolveFixture{
		Kind:      "grok",
		SessionID: fixtureGrokSessionID,
		Tree: []iterm2.AgentTreeNode{
			{PID: 200, PPID: 1, Role: "input", Cmd: "/usr/local/bin/agent-run run --session-id=ignored-cli"},
			{PID: 201, PPID: 200, Role: "agent-run-serve", Cmd: "/usr/local/bin/agent-run serve --session-id=ignored-cli"},
			{PID: 202, PPID: 201, Role: "grok", Cmd: "/usr/local/bin/grok"},
		},
	}
}

func installFixture(t *testing.T, req *Request) {
	t.Helper()
	running := true
	if req.ITermRunning != nil {
		running = *req.ITermRunning
	}
	opts := iterm2.PhasedFixtureOpts{
		Windows:           fixtureWindows(),
		ITermRunning:      running,
		IdleTTYs:          []string{"ttys001"},
		BusyTTYs:          []string{"ttys002"},
		Now:               time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname:          "testhost",
		AgentResolveByTTY: req.AgentResolveByTTY,
	}
	iterm2.InstallPhasedFixtureCollectorForTest(t, opts)
}

// Run exercises sessions snapshot / session status / help with inject.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	resp := &Response{}

	switch req.Mode {
	case ModeHelp:
		var stdout, stderr bytes.Buffer
		args := req.HelpArgs
		if len(args) == 0 {
			args = []string{"sessions", "-h"}
		}
		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.ExitCode = code
		return resp, nil

	case ModeSnapshotCLI:
		if req.UseTwoWindowFixture || len(req.AgentResolveByTTY) > 0 || req.ITermRunning != nil {
			installFixture(t, req)
		}
		var stdout, stderr bytes.Buffer
		args := []string{"sessions", "snapshot"}
		if req.JSON {
			args = append(args, "--json")
		}
		if req.NoColor {
			args = append(args, "--no-color")
		}
		if req.NoStream {
			args = append(args, "--no-stream")
		}
		if req.NoEnrich {
			args = append(args, "--no-enrich")
		}
		if req.NoTree {
			args = append(args, "--no-tree")
		}
		args = append(args, req.ExtraArgs...)
		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.ExitCode = code
		return resp, nil

	case ModeSessionStatus:
		if req.UseTwoWindowFixture || len(req.AgentResolveByTTY) > 0 || req.ITermRunning != nil {
			installFixture(t, req)
		}
		ref := req.SessionRef
		if ref == "" {
			ref = fixtureITermBusyID
		}
		var stdout, stderr bytes.Buffer
		args := []string{"session", ref, "status"}
		if req.JSON {
			args = append(args, "--json")
		}
		if req.NoColor {
			args = append(args, "--no-color")
		}
		if req.NoEnrich {
			args = append(args, "--no-enrich")
		}
		if req.NoTree {
			args = append(args, "--no-tree")
		}
		args = append(args, req.ExtraArgs...)
		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.ExitCode = code
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// silence unused imports used only in ASSERT leaves.
var (
	_ = json.Valid
	_ = os.TempDir
	_ = filepath.Join
)
```
