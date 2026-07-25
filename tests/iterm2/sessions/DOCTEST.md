# kool iterm2 sessions snapshot — multi-phase AS + streaming CLI (P3)

`kool iterm2 sessions snapshot` captures live iTerm2 windows, tabs, and sessions
for human review. **P3** splits hierarchy collection into composable AppleScript
phases and streams CLI window blocks as soon as each window is fully known
(with process idle/busy enrich). Structured formats stay fully buffered.

**Out of scope for this tree:** procresolve / grok / codex session lines, pid
tree display, agent-pro dependency.

Sibling trees: open-dir `./tests/iterm2`, tab-set `./tests/iterm2/tab-set`.

## Version

0.0.2

**Classic TDD (P3):** phased APIs and streaming CLI leaves are intentionally
**RED** until the implementer lands multi-phase collection + stream. Help /
format-conflict can approach **GREEN** on existing CLI without new symbols;
leaves that need fixture inject stay **RED** until the test helper + phased
API surface compile and behave.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — invokes `kool iterm2 sessions …` or package APIs under test.
- **kool CLI / handler** — `tools/iterm2` routes `sessions` → snapshot; format
  flags, `-o`, and **`--no-stream`**.
- **SnapshotCollector** — injectable via `SetSnapshotCollectorForTest`. **P3**
  exposes phased hierarchy methods and uses them inside `Capture`.
- **Phased AppleScript** — composable queries (not one monolithic dump):
  - `ListWindows()` — window index + name
  - `ListTabsAndSessions(windowIndex)` — tabs + sessions for one window
- **Process enricher** — per-session `ps`/cwd → idle/busy (existing; no agent).
- **CLI streamer** — default CLI emits each window block after that window is
  collected **and** enriched; footer summary last. JSON / Markdown / HTML /
  `-o` / `--no-stream` collect fully, then render once.
- **Fixture installer (tests)** — `InstallPhasedFixtureCollectorForTest` so
  external doctests never touch unexported `rawProc` or live iTerm.

### Behaviors

**Commands**

```text
kool iterm2 sessions -h|--help
kool iterm2 sessions snapshot [options]
kool iterm2 sessions snapshot -h|--help
```

**Snapshot options (P3)**

| Flag / mode | Effect |
|-------------|--------|
| default CLI | Stream window blocks progressively; footer last |
| `--no-stream` | Buffer full CLI (collect all → render once) |
| `--json` / `--markdown` / `--html` | Always one buffered document |
| `-o/--output FILE` | Buffered file write; stderr `Wrote …` |
| `--no-color` | No ANSI on CLI |
| multiple format flags | error “mutually exclusive”, exit 1 |

**Expected implementer package surface (tools/iterm2)**

```go
// Phased hierarchy (methods on *SnapshotCollector; Capture uses them):
func (c *SnapshotCollector) ListWindows() (windows []SnapshotWindow, warnings []string, err error)
func (c *SnapshotCollector) ListTabsAndSessions(windowIndex int) (tabs []SnapshotTab, warnings []string, err error)

// Optional test accessor:
func ActiveSnapshotCollectorForTest() *SnapshotCollector

// Doctest fixture helper (package iterm2; uses unexported rawProc internally):
type PhasedFixtureOpts struct {
    Windows       []SnapshotWindow
    ITermRunning  bool
    OnListWindows func()
    OnListTabs    func(windowIndex int) // called at start of each ListTabsAndSessions
    IdleTTYs      []string              // short names e.g. "ttys001"
    BusyTTYs      []string
    Now           time.Time
    Hostname      string
}
func InstallPhasedFixtureCollectorForTest(t testing.TB, opts PhasedFixtureOpts)

// CLI: parse --no-stream; stream default CLI only when format is CLI and !NoStream.
```

**Streaming order (observable contract)**

1. `ListWindows`
2. For window index 1..N: `ListTabsAndSessions(i)` → enrich sessions → **write
   CLI block including `W{i}`** to stdout
3. Footer summary after all windows

When `ListTabsAndSessions` for the **last** window begins, default streaming
CLI must already have written `W1` to stdout. With `--no-stream` / JSON / etc.,
`W1` must **not** appear until collection finishes (probe sees no `W1` at last
ListTabs start).

**Errors**

- iTerm not running → non-zero; message mentions iTerm2 / not running
- format flag conflict → mutually exclusive
- unexpected args → exit 1

### Inject / no live iTerm

`Run` installs the fixture collector (except pure help / format-conflict).
Never requires a real iTerm2 process.

## Decision Tree

```
sessions/                               [nested DOCTEST root — P3]
├── help/
│   └── show-usage/                     snapshot, --json, --no-stream on help
├── phased/                             [composable hierarchy APIs]
│   ├── list-windows/                   ListWindows → Win-A, Win-B headers
│   ├── list-tabs-and-sessions/         ListTabsAndSessions(1) → Tab-A + session
│   └── capture-uses-phased/            Capture: 1× ListWindows + N× ListTabs
├── stream/                             [CLI progressive vs buffer]
│   ├── order/                          W1 before last-window ListTabs
│   ├── no-stream/                      --no-stream: full CLI; no progressive W1
│   └── enrich-before-emit/             streamed lines include idle/busy
├── format/                             [buffered structured formats]
│   ├── json-buffered/                  --json one valid JSON object
│   ├── markdown-buffered/              --markdown full doc
│   ├── html-buffered/                  --html contains <html
│   └── output-file/                    -o file; stderr Wrote
└── error/
    ├── iterm-not-running/              not running → non-zero
    └── format-conflict/                --json --html mutually exclusive
```

## Test Index

| Leaf | Description | Expect |
|------|-------------|--------|
| `help/show-usage/` | Help mentions `snapshot`, `--json`, `--no-stream` | RED until `--no-stream` documented |
| `phased/list-windows/` | ListWindows returns two fixture windows | RED |
| `phased/list-tabs-and-sessions/` | ListTabsAndSessions(1) returns tabs/sessions | RED |
| `phased/capture-uses-phased/` | Capture: 1 ListWindows + ListTabs per window | RED |
| `stream/order/` | Default CLI: W1 before last ListTabs | RED |
| `stream/no-stream/` | `--no-stream` full CLI + footer; not progressive | RED |
| `stream/enrich-before-emit/` | Streamed output includes idle and busy | RED |
| `format/json-buffered/` | `--json` single JSON document | RED until fixture helper; then GREEN |
| `format/markdown-buffered/` | `--markdown` buffered | RED until fixture helper; then GREEN |
| `format/html-buffered/` | `--html` buffered | RED until fixture helper; then GREEN |
| `format/output-file/` | `-o` Wrote + file content | RED until fixture helper; then GREEN |
| `error/iterm-not-running/` | not running → non-zero | RED until fixture helper; then GREEN |
| `error/format-conflict/` | `--json --html` → mutually exclusive | GREEN (existing) |

## How to Run

```sh
doctest vet ./tests/iterm2/sessions
doctest test ./tests/iterm2/sessions
```

Nested root: no inheritance from `./tests/iterm2` open-dir `Request`/`Run`.
Designer does **not** implement production streaming code.

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

// Mode selects which surface Run exercises.
const (
	ModeHelp          = "help"
	ModeListWindows   = "list-windows"
	ModeListTabs      = "list-tabs"
	ModeCapturePhased = "capture-phased"
	ModeSnapshotCLI   = "snapshot-cli"
)

// Request drives one in-process sessions/snapshot or phased API scenario.
type Request struct {
	// Mode is help | list-windows | list-tabs | capture-phased | snapshot-cli.
	Mode string

	// HelpArgs when ModeHelp (default: sessions -h).
	HelpArgs []string

	// Snapshot CLI flags (ModeSnapshotCLI).
	JSON       bool
	Markdown   bool
	HTML       bool
	NoColor    bool
	NoStream   bool
	OutputPath string // relative to WorkingDir unless absolute
	ExtraArgs  []string

	// UseTwoWindowFixture installs the canonical two-window inject.
	UseTwoWindowFixture bool

	// ObserveStreamOrder records SawW1BeforeLastListTabs via OnListTabs probe.
	ObserveStreamOrder bool

	// ITermRunning nil → true; false → not-running error path.
	ITermRunning *bool

	WorkingDir string
}

// Response is capture + inject observations after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int

	Windows  []iterm2.SnapshotWindow
	Tabs     []iterm2.SnapshotTab
	Snapshot *iterm2.Snapshot
	Warnings []string
	APIError string

	ListWindowsCalls int
	ListTabsCalls    int
	ListTabsByIndex  map[int]int

	// SawW1BeforeLastListTabs: stdout already contained "W1" when ListTabs for
	// window index 2 started (streaming contract).
	SawW1BeforeLastListTabs bool
	LastListTabsInvoked     bool

	OutputFile string
}

func boolPtr(v bool) *bool { return &v }

// fixtureWindows is the canonical two-window P3 hierarchy fixture.
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
							ID:      "AAAAAAAA-0000-0000-0000-000000000001",
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
							ID:      "BBBBBBBB-0000-0000-0000-000000000002",
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

// streamProbe observes phased call counts and progressive W1 emission.
type streamProbe struct {
	getStdout func() string

	listWindows int
	listTabs    int
	byIndex     map[int]int
	sawW1       bool
	lastInvoked bool
	lastIndex   int
}

func newStreamProbe(getStdout func() string, lastWindowIndex int) *streamProbe {
	return &streamProbe{
		getStdout: getStdout,
		byIndex:   map[int]int{},
		lastIndex: lastWindowIndex,
	}
}

func (p *streamProbe) onListWindows() {
	if p == nil {
		return
	}
	p.listWindows++
}

func (p *streamProbe) onListTabs(windowIndex int) {
	if p == nil {
		return
	}
	p.listTabs++
	p.byIndex[windowIndex]++
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
	resp.ListWindowsCalls = p.listWindows
	resp.ListTabsCalls = p.listTabs
	resp.ListTabsByIndex = p.byIndex
	resp.SawW1BeforeLastListTabs = p.sawW1
	resp.LastListTabsInvoked = p.lastInvoked
}

func installFixture(t *testing.T, req *Request, probe *streamProbe) {
	t.Helper()
	running := true
	if req.ITermRunning != nil {
		running = *req.ITermRunning
	}
	opts := iterm2.PhasedFixtureOpts{
		Windows:      fixtureWindows(),
		ITermRunning: running,
		IdleTTYs:     []string{"ttys001"},
		BusyTTYs:     []string{"ttys002"},
		Now:          time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname:     "testhost",
	}
	if probe != nil {
		opts.OnListWindows = probe.onListWindows
		opts.OnListTabs = probe.onListTabs
	}
	iterm2.InstallPhasedFixtureCollectorForTest(t, opts)
}

// Run exercises phased APIs or sessions snapshot CLI in-process with inject.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	resp := &Response{ListTabsByIndex: map[int]int{}}

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

	case ModeListWindows:
		probe := newStreamProbe(nil, 2)
		installFixture(t, req, probe)
		c := iterm2.ActiveSnapshotCollectorForTest()
		wins, warns, err := c.ListWindows()
		resp.Windows = wins
		resp.Warnings = warns
		probe.applyTo(resp)
		if err != nil {
			resp.APIError = err.Error()
			resp.ExitCode = 1
		}
		return resp, nil

	case ModeListTabs:
		probe := newStreamProbe(nil, 2)
		installFixture(t, req, probe)
		c := iterm2.ActiveSnapshotCollectorForTest()
		tabs, warns, err := c.ListTabsAndSessions(1)
		resp.Tabs = tabs
		resp.Warnings = warns
		probe.applyTo(resp)
		if err != nil {
			resp.APIError = err.Error()
			resp.ExitCode = 1
		}
		return resp, nil

	case ModeCapturePhased:
		probe := newStreamProbe(nil, 2)
		installFixture(t, req, probe)
		snap, warns, err := iterm2.CaptureSnapshot()
		resp.Snapshot = snap
		resp.Warnings = warns
		probe.applyTo(resp)
		if err != nil {
			resp.APIError = err.Error()
			resp.ExitCode = 1
		}
		return resp, nil

	case ModeSnapshotCLI:
		var stdout, stderr bytes.Buffer
		probe := newStreamProbe(func() string { return stdout.String() }, 2)

		needFixture := req.UseTwoWindowFixture || req.ObserveStreamOrder || req.ITermRunning != nil
		// Format / stream leaves always use the two-window fixture.
		if needFixture || req.JSON || req.Markdown || req.HTML || req.NoStream || req.OutputPath != "" || req.NoColor {
			// Skip fixture only when pure format-conflict (no capture).
			if !(req.JSON && req.HTML) {
				installFixture(t, req, probe)
			}
		}

		args := []string{"sessions", "snapshot"}
		if req.JSON {
			args = append(args, "--json")
		}
		if req.Markdown {
			args = append(args, "--markdown")
		}
		if req.HTML {
			args = append(args, "--html")
		}
		if req.NoColor {
			args = append(args, "--no-color")
		}
		if req.NoStream {
			args = append(args, "--no-stream")
		}
		outPath := req.OutputPath
		if outPath != "" {
			if !filepath.IsAbs(outPath) {
				outPath = filepath.Join(req.WorkingDir, outPath)
			}
			args = append(args, "-o", outPath)
		}
		args = append(args, req.ExtraArgs...)

		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.ExitCode = code
		probe.applyTo(resp)

		if outPath != "" {
			if b, err := os.ReadFile(outPath); err == nil {
				resp.OutputFile = string(b)
			}
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// silence unused import if a leaf only needs encoding/json in Assert.
var _ = json.Valid
```
