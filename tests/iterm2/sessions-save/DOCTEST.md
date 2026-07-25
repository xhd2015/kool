# kool iterm2 sessions save / restore

Checkpoint critical **grok**, **codex**, and **mark** tabs to
`~/.config/iterm2/sessions-save.json` (override with `--file`), then restore
windows with `cd` + resume commands.

**Sibling of** `./tests/iterm2/sessions` (snapshot) and `sessions-p4` (enrich).

## Version

0.0.2

**Classic TDD (color + streaming):** dry-run color flags and progressive save
dry-run stream are intentionally **RED** until the implementer lands
`--color`/`--no-color`, colored plan tokens, and per-window stream emit.
Existing dry-run / write / restore leaves stay contracted and may remain
**GREEN** on monochrome buffered output.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — invokes `kool iterm2 sessions save|restore …`.
- **kool CLI / handler** — `tools/iterm2` routes `sessions save` / `sessions restore`;
  flags `--dry-run`, `--file`, **`--color`**, **`--no-color`**, `-h/--help`.
- **SnapshotCollector** — injectable via `InstallPhasedFixtureCollectorForTest`;
  phased `ListWindows` / `ListTabsAndSessions` + enrich + agent resolve.
- **Critical filter** — keeps grok/codex (session_id) and live mark panes;
  builds `SaveDocument` (JSON shape unchanged).
- **Save plan streamer** — on save `--dry-run`, after each window is captured
  and classified with ≥1 critical tab, writes that **window block** to stdout
  immediately; after all windows, writes the **footer** (Would save + path +
  dry-run note). No “scanning…” header. Live save may share the same path
  (window plan while capturing; file write after).
- **Restore planner** — restore `--dry-run` prints header → each window/tab →
  footer line-at-a-time (no full-buffer dump).
- **Color resolver** — same policy as snapshot / go-best-practice cli/color:
  `--color` && `--no-color` → error; `--color` force on; `--no-color` force
  off; else `NO_COLOR` non-empty → off; else stdout TTY.
- **Fixture installer (tests)** — phased fixture + optional `OnListTabs` stream
  probe; no live iTerm.

### Critical filter

| Kind | Detection | Resume |
|------|-----------|--------|
| grok | `agent.kind=grok` + `session_id` | `grok --resume <id>` |
| codex | `agent.kind=codex` + `session_id` | `codex resume <id>` |
| mark | live process basename `mark` | `mark` / `mark '<message>'` |

Agent preferred over mark on the same pane. Empty cwd → skip + warning.
Zero critical → exit 0, **no write**.

### File lifecycle

- `saved_at` set on write; `restored_at` null until restore succeeds
- save when pending + non-TTY → error
- save when already restored → overwrite without prompt
- restore when `restored_at` set → error (consumed)
- **No ANSI in checkpoint JSON**

### Color tokens (stdout)

| Token | Use |
|-------|-----|
| Green | `Would save`, `Would restore`, `Saved`, `Restored` |
| Bold | `W{n}`, `new window` |
| Green kinds | `grok`, `codex` |
| Yellow kind | `mark` |
| Gray | path, cwd, resume_cmd lines, counts meta, `(dry-run: not written\|not applied)`, saved_at meta |

Stderr: existing yellow `warning:`, red `Error:` (unchanged).

### Streaming order (save --dry-run, observable)

1. For each window after enrich: classify critical tabs
2. If ≥1 critical → **write that window block** (`W{n}` …) to stdout immediately
3. After all windows → footer: green Would save + gray path + dry-run note

When `ListTabsAndSessions` for the **last** window begins and ≥2 windows have
critical content, stdout must already contain `W1` (stream probe).

### Inject

- `InstallPhasedFixtureCollectorForTest` + `AgentResolveByTTY` / `BusyLeafByTTY`
- Restore AppleScript via package hook (unit tests); doctest focuses on
  dry-run + file IO + help + color/stream contracts

## Decision Tree

```
sessions-save/
├── help/
│   ├── show-usage/           sessions -h mentions save + restore
│   └── color-flags/          save -h / restore -h mention --color + --no-color
├── save/
│   ├── dry-run/              plan only; no file (auto color; pipe OK monochrome)
│   ├── write/                writes version + saved_at; no restored_at
│   ├── zero/                 0 critical; no write
│   ├── pending-non-tty/      existing pending + non-TTY → error
│   ├── color/
│   │   ├── force-on/         save --dry-run --color → ANSI (green/bold)
│   │   ├── force-off/        --no-color → no ESC
│   │   └── conflict/         --color --no-color → non-zero + together msg
│   └── stream/
│       └── order/            two critical windows; W1 before last ListTabs
└── restore/
    ├── dry-run/              plan shows cd + resume; not stamped
    ├── consumed/             restored_at set → error
    └── color/
        └── force-on/         restore --dry-run --color → ANSI
```

## Test Index

| Leaf | Description | Expect |
|------|-------------|--------|
| `help/show-usage/` | Help mentions `save`, `restore`, `snapshot` | GREEN |
| `help/color-flags/` | save + restore help document `--color` / `--no-color` | RED until flags documented |
| `save/dry-run/` | Would save + critical ids; no file | GREEN (monochrome OK) |
| `save/write/` | file version, saved_at, kinds | GREEN |
| `save/zero/` | 0 critical; no write | GREEN |
| `save/pending-non-tty/` | pending + non-TTY error | GREEN |
| `save/color/force-on/` | `--color` ANSI on plan | RED |
| `save/color/force-off/` | `--no-color` no `\x1b` | RED until flag accepted (then GREEN if monochrome default) |
| `save/color/conflict/` | both flags → cannot be specified together | RED |
| `save/stream/order/` | W1 before last ListTabs | RED |
| `restore/dry-run/` | Would restore; not stamped | GREEN |
| `restore/consumed/` | consumed error | GREEN |
| `restore/color/force-on/` | restore `--color` ANSI | RED |

## How to Run

```sh
doctest vet ./tests/iterm2/sessions-save
doctest test ./tests/iterm2/sessions-save
```

```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

const (
	ModeHelp    = "help"
	ModeSave    = "save"
	ModeRestore = "restore"
)

const (
	fixtureGrokSessionID = "019fabcdef-1234-5678-9abc-def012345678"
)

type Request struct {
	Mode string

	HelpArgs []string
	// HelpCombinedSaveRestore: ModeHelp captures save -h then restore -h into Stdout.
	HelpCombinedSaveRestore bool

	DryRun    bool
	FilePath  string // absolute or relative to WorkingDir
	ExtraArgs []string

	// Color force flags (mapped to --color / --no-color on save|restore).
	Color   bool
	NoColor bool

	// Install critical fixture (one window: grok + mark + idle).
	UseCriticalFixture bool
	// Idle-only fixture (zero critical).
	UseIdleOnlyFixture bool
	// Two windows each with ≥1 critical tab (stream order).
	UseTwoCriticalWindows bool

	// ObserveStreamOrder records SawW1BeforeLastListTabs via OnListTabs probe.
	ObserveStreamOrder bool

	// Pre-seed checkpoint at FilePath before run.
	SeedDoc *iterm2.SaveDocument

	// Force non-TTY for overwrite checks (default true for save tests).
	NonTTY *bool

	WorkingDir string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	FileJSON string
	Doc      *iterm2.SaveDocument

	// Stream probe (save path with ObserveStreamOrder).
	SawW1BeforeLastListTabs bool
	LastListTabsInvoked     bool
	ListTabsCalls           int
}

func boolPtr(v bool) *bool { return &v }

func criticalWindows() []iterm2.SnapshotWindow {
	return []iterm2.SnapshotWindow{
		{
			Index: 1,
			Name:  "Win-Crit",
			Tabs: []iterm2.SnapshotTab{
				{Index: 1, Sessions: []iterm2.SnapshotSession{
					{Index: 1, ID: "AAAAAAAA-0000-0000-0000-000000000001", TTY: "/dev/ttys001", Profile: "Default"},
				}},
				{Index: 2, Sessions: []iterm2.SnapshotSession{
					{Index: 1, ID: "BBBBBBBB-0000-0000-0000-000000000002", TTY: "/dev/ttys002", Profile: "Default"},
				}},
				{Index: 3, Sessions: []iterm2.SnapshotSession{
					{Index: 1, ID: "CCCCCCCC-0000-0000-0000-000000000003", TTY: "/dev/ttys003", Profile: "Default"},
				}},
			},
		},
	}
}

// twoCriticalWindows is the stream-order fixture: W1 grok, W2 mark (both critical).
func twoCriticalWindows() []iterm2.SnapshotWindow {
	return []iterm2.SnapshotWindow{
		{
			Index: 1,
			Name:  "Win-A",
			Tabs: []iterm2.SnapshotTab{
				{Index: 1, Sessions: []iterm2.SnapshotSession{
					{Index: 1, ID: "AAAAAAAA-0000-0000-0000-000000000001", TTY: "/dev/ttys001", Profile: "Default"},
				}},
			},
		},
		{
			Index: 2,
			Name:  "Win-B",
			Tabs: []iterm2.SnapshotTab{
				{Index: 1, Sessions: []iterm2.SnapshotSession{
					{Index: 1, ID: "BBBBBBBB-0000-0000-0000-000000000002", TTY: "/dev/ttys002", Profile: "Default"},
				}},
			},
		},
	}
}

// streamProbe observes progressive W1 emission at last-window ListTabs start.
type streamProbe struct {
	getStdout   func() string
	listTabs    int
	sawW1       bool
	lastInvoked bool
	lastIndex   int
}

func newStreamProbe(getStdout func() string, lastWindowIndex int) *streamProbe {
	return &streamProbe{
		getStdout: getStdout,
		lastIndex: lastWindowIndex,
	}
}

func (p *streamProbe) onListTabs(windowIndex int) {
	if p == nil {
		return
	}
	p.listTabs++
	if windowIndex == p.lastIndex {
		p.lastInvoked = true
		if p.getStdout != nil && strings.Contains(p.getStdout(), "W1") {
			p.sawW1 = true
		}
	}
}

func (p *streamProbe) applyTo(resp *Response) {
	if p == nil {
		return
	}
	resp.ListTabsCalls = p.listTabs
	resp.SawW1BeforeLastListTabs = p.sawW1
	resp.LastListTabsInvoked = p.lastInvoked
}

func installCritical(t *testing.T, probe *streamProbe) {
	t.Helper()
	opts := iterm2.PhasedFixtureOpts{
		ITermRunning: true,
		Windows:      criticalWindows(),
		BusyTTYs:     []string{"ttys001", "ttys002"},
		IdleTTYs:     []string{"ttys003"},
		BusyLeafByTTY: map[string]string{
			"ttys002": "mark still waiting for CI",
		},
		CwdByTTY: map[string]string{
			"ttys001": "/proj/grok",
			"ttys002": "/proj/mark",
		},
		AgentResolveByTTY: map[string]iterm2.AgentResolveFixture{
			"ttys001": {Kind: "grok", SessionID: fixtureGrokSessionID, Title: "fixture-title"},
		},
		Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname: "testhost",
	}
	if probe != nil {
		opts.OnListTabs = probe.onListTabs
	}
	iterm2.InstallPhasedFixtureCollectorForTest(t, opts)
}

func installTwoCritical(t *testing.T, probe *streamProbe) {
	t.Helper()
	opts := iterm2.PhasedFixtureOpts{
		ITermRunning: true,
		Windows:      twoCriticalWindows(),
		BusyTTYs:     []string{"ttys001", "ttys002"},
		BusyLeafByTTY: map[string]string{
			"ttys002": "mark still waiting for CI",
		},
		CwdByTTY: map[string]string{
			"ttys001": "/proj/a",
			"ttys002": "/proj/b",
		},
		AgentResolveByTTY: map[string]iterm2.AgentResolveFixture{
			"ttys001": {Kind: "grok", SessionID: fixtureGrokSessionID, Title: "fixture-a"},
		},
		Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname: "testhost",
	}
	if probe != nil {
		opts.OnListTabs = probe.onListTabs
	}
	iterm2.InstallPhasedFixtureCollectorForTest(t, opts)
}

func installIdleOnly(t *testing.T) {
	t.Helper()
	iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []iterm2.SnapshotWindow{{
			Index: 1,
			Tabs: []iterm2.SnapshotTab{{Index: 1, Sessions: []iterm2.SnapshotSession{
				{Index: 1, ID: "AAAAAAAA-0000-0000-0000-000000000001", TTY: "/dev/ttys001", Profile: "Default"},
			}}},
		}},
		IdleTTYs: []string{"ttys001"},
		Hostname: "testhost",
	})
}

func resolveFile(req *Request) string {
	p := req.FilePath
	if p == "" {
		p = "sessions-save.json"
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(req.WorkingDir, p)
	}
	return p
}

func appendColorFlags(args []string, req *Request) []string {
	if req.Color {
		args = append(args, "--color")
	}
	if req.NoColor {
		args = append(args, "--no-color")
	}
	return args
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	path := resolveFile(req)

	if req.SeedDoc != nil {
		if err := iterm2.WriteSaveDocument(path, req.SeedDoc); err != nil {
			return nil, err
		}
	}

	nonTTY := true
	if req.NonTTY != nil {
		nonTTY = *req.NonTTY
	}
	// Non-TTY is the default for CI; real TTY path is unit-tested separately.
	_ = nonTTY

	switch req.Mode {
	case ModeHelp:
		var stdout, stderr bytes.Buffer
		if req.HelpCombinedSaveRestore {
			var code int
			for _, args := range [][]string{
				{"sessions", "save", "-h"},
				{"sessions", "restore", "-h"},
			} {
				var so, se bytes.Buffer
				c := iterm2.RunForTest(args, &so, &se, req.WorkingDir)
				stdout.WriteString(so.String())
				if !strings.HasSuffix(so.String(), "\n") {
					stdout.WriteByte('\n')
				}
				stderr.WriteString(se.String())
				if c != 0 {
					code = c
				}
			}
			return &Response{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}, nil
		}
		args := req.HelpArgs
		if len(args) == 0 {
			args = []string{"sessions", "-h"}
		}
		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		return &Response{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}, nil

	case ModeSave:
		var stdout, stderr bytes.Buffer
		var probe *streamProbe
		if req.ObserveStreamOrder {
			probe = newStreamProbe(func() string { return stdout.String() }, 2)
		}

		if req.UseIdleOnlyFixture {
			installIdleOnly(t)
		} else if req.UseTwoCriticalWindows {
			installTwoCritical(t, probe)
		} else if req.UseCriticalFixture {
			installCritical(t, probe)
		}

		args := []string{"sessions", "save", "--file", path}
		if req.DryRun {
			args = append(args, "--dry-run")
		}
		args = appendColorFlags(args, req)
		args = append(args, req.ExtraArgs...)
		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}
		if probe != nil {
			probe.applyTo(resp)
		}
		if b, err := os.ReadFile(path); err == nil {
			resp.FileJSON = string(b)
			var doc iterm2.SaveDocument
			if json.Unmarshal(b, &doc) == nil {
				resp.Doc = &doc
			}
		}
		return resp, nil

	case ModeRestore:
		var stdout, stderr bytes.Buffer
		args := []string{"sessions", "restore", "--file", path}
		if req.DryRun {
			args = append(args, "--dry-run")
		}
		args = appendColorFlags(args, req)
		args = append(args, req.ExtraArgs...)
		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}
		if b, err := os.ReadFile(path); err == nil {
			resp.FileJSON = string(b)
			var doc iterm2.SaveDocument
			if json.Unmarshal(b, &doc) == nil {
				resp.Doc = &doc
			}
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}
```
