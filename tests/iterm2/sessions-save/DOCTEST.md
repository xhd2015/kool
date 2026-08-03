# kool iterm2 sessions save / restore

Checkpoint critical **grok**, **codex**, and **mark** tabs to
`~/.config/iterm2/sessions-save.json` (override with `--file`), then restore
windows with `cd` + resume commands. **P2** also records macOS Space
(`space` + `iterm_window_id`) and restores placement by Space index.

**Sibling of** `./tests/iterm2/sessions` (snapshot) and `sessions-p4` (enrich).

## Version

0.0.3

**Classic TDD (color + streaming + Space + already-running):** dry-run color
flags, progressive save stream, **macOS Space**, and **restore already-running
skip** leaves are intentionally **RED** until the implementer lands those
behaviors. Existing dry-run / write / restore / help leaves stay contracted and
should remain **GREEN** on monochrome buffered output (zero-value Request
defaults — no Space / already-running fixture flags). After already-running
lands, restore always attempts a live capture; capture fail is soft (D6) so
prior leaves without fixtures stay GREEN.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — invokes `kool iterm2 sessions save|restore …`.
- **kool CLI / handler** — `tools/iterm2` routes `sessions save` / `sessions restore`;
  flags `--dry-run`, `--file`, **`--color`**, **`--no-color`**,
  **`--ignore-macos-space`**, `-h/--help`.
- **SnapshotCollector** — injectable via `InstallPhasedFixtureCollectorForTest`;
  phased `ListWindows` / `ListTabsAndSessions` + enrich + agent resolve.
  Fixtures may carry per-window **WindowID** (iTerm/CG window number) once the
  model gains that field.
- **Space index resolver** — P1 `space.SpaceIndexForWindow(windowID)` (via go.mod
  replace to `external/dot-pkgs-…/go-pkgs`). Save maps window id → 0-based
  Desktop index. Injectable for tests (implementer hook; no live WindowServer).
- **Space Backend** — Create / Switch / Highest for restore placement
  (`MockBackend` pattern from `computer-use/macos/space`). Live restore only;
  dry-run never calls it.
- **Critical filter** — keeps grok/codex (session_id) and live mark panes;
  builds `SaveDocument`.
- **Save plan streamer** — on save `--dry-run`, after each window is captured
  and classified with ≥1 critical tab, writes that **window block** to stdout
  immediately (includes **space N (Desktop N+1)** meta when Space recording is
  on); after all windows, writes the **footer**. No “scanning…” header.
- **Restore planner** — restore `--dry-run` prints header → each window (with
  space meta) / tab → footer. No Switch/Create/AS side effects for placement.
- **Already-running scanner** — after valid unconsumed checkpoint load (dry-run
  and live), capture live snapshot (enrich on), index critical panes by
  agent `kind`+`session_id` and mark exact `message`. Checkpoint tabs that hit
  a live pane are **skipped** (no create/resume); stderr warning per hit
  (first live hit only). Header counts = would-create only; skip count in meta
  when `skipped > 0`. Capture fail → soft warning, 0 hits, restore all.
- **Color resolver** — same policy as snapshot / go-best-practice cli/color:
  `--color` && `--no-color` → error; `--color` force on; `--no-color` force
  off; else `NO_COLOR` non-empty → off; else stdout TTY.
- **Fixture installer (tests)** — phased fixture + optional `OnListTabs` stream
  probe; `SeedRawJSON` for checkpoints with `space` / `iterm_window_id` without
  requiring production struct fields yet. No live iTerm / Mission Control.

### Critical filter

| Kind | Detection | Resume |
|------|-----------|--------|
| grok | `agent.kind=grok` + `session_id` | `grok --resume <id>` |
| codex | `agent.kind=codex` + `session_id` | `codex resume <id>` |
| mark | live process basename `mark` | `mark` / `mark '<message>'` |

Agent preferred over mark on the same pane. Empty cwd → skip + warning.
Zero critical → exit 0, **no write**.

### Checkpoint Space fields (`SaveWindow`)

| Field | Rule |
|-------|------|
| `space` | 0-based Desktop index. When **not** `--ignore-macos-space`, **always emit** including `0`. Missing on read → 0. |
| `iterm_window_id` | iTerm AS / CG window number at save; **info only**. Restore never uses it. Emitted when known. |
| `--ignore-macos-space` on save | Omit **both** `space` and `iterm_window_id`. No resolve. |
| `--ignore-macos-space` on restore | Ignore `space`; create windows on current Desktop (legacy AS path). No Switch/Create. |

### Save (Space)

1. Snapshot captures per-window iTerm window id (fixture `WindowID` / AS `id of window`).
2. Unless ignore: resolve `space` via `SpaceIndexForWindow(iterm_window_id)` (injectable).
3. Resolve fail / non-type-0 → `space=0` + **warning** (stderr yellow); still may emit `iterm_window_id` if known.
4. Always emit `space` when not ignore (including 0).

### Restore (Space)

1. Unless ignore, for each window in order:
   - `s = space` (default 0 if missing); if `s >= 16` → warn, `s=0`
   - if `s==0`: **Switch(Desktop 1)** always (when placement on)
   - else: Create until Highest >= s+1; on max-cap fail → warn, s=0, Switch(1); else Switch(s+1)
   - create that window+tabs (per-window AS)
2. Never use `iterm_window_id` for placement.
3. Dry-run: show `space N (Desktop N+1)` only — **not** `iterm_window_id`.
   Clamp `s>=16` applies to the **planned** label (`space 0 (Desktop 1)`) and
   emits a **warning on stderr** even in dry-run (no Switch/Create).
4. `--ignore-macos-space`: dry-run omits space placement lines; no Switch/Create.

### Already-running skip (restore)

| Rule | Behavior |
|------|----------|
| Match key | agent: `kind` + `session_id`; mark: exact `message` |
| On hit | **skip** tab (no create/resume); warn on stderr |
| Warning shape | `tab "<name>" (kind id) is already running at space N (Desktop N+1), pid P` (name optional; space soft) |
| Multi-hit | first live hit only |
| Dry-run | full plan: skip markers + would-restore for remaining; not stamped; header = would-create counts; skip meta when skipped>0 |
| Live | AS only non-skipped tabs; all-skipped windows omitted; all remaining 0 → still stamp `restored_at`, exit 0, no AS create |
| Capture fail | soft warn; treat as 0 hits; restore all |
| Flags / schema | no new CLI flags or checkpoint JSON fields |

### File lifecycle

- `saved_at` set on write; `restored_at` null until restore succeeds
- save when pending + non-TTY → error
- save when already restored → overwrite without prompt
- restore when `restored_at` set → error (consumed)
- all tabs skipped on live restore → still stamp `restored_at` (E1)
- **No ANSI in checkpoint JSON**

### Color tokens (stdout)

| Token | Use |
|-------|-----|
| Green | `Would save`, `Would restore`, `Saved`, `Restored` |
| Bold | `W{n}`, `new window` |
| Green kinds | `grok`, `codex` |
| Yellow kind | `mark` |
| Gray | path, cwd, resume_cmd lines, counts meta, space meta, `(dry-run: not written\|not applied)`, saved_at meta |

Stderr: existing yellow `warning:`, red `Error:` (unchanged).

### Streaming order (save --dry-run, observable)

1. For each window after enrich: classify critical tabs
2. If ≥1 critical → **write that window block** (`W{n}` …) to stdout immediately
3. After all windows → footer: green Would save + gray path + dry-run note

When `ListTabsAndSessions` for the **last** window begins and ≥2 windows have
critical content, stdout must already contain `W1` (stream probe).

### Inject (tests)

- `InstallPhasedFixtureCollectorForTest` + `AgentResolveByTTY` / `BusyLeafByTTY`
- **Implementer surface (expected):**
  - `SnapshotWindow.WindowID` (uint64) for fixture `iterm_window_id`
  - `SaveWindow.Space` / `SaveWindow.ItermWindowID` (JSON `space` / `iterm_window_id`)
  - test hook to inject Space index resolver (fixed index / error)
  - test hook to inject Space Backend + restore AppleScript for live placement
  - already-running scan after checkpoint load (dry-run + live); soft capture fail
- Doctest harness: `SeedRawJSON` / `SeedDoc`, `--ignore-macos-space` via Request,
  `RestoreLiveFixture` + `FailSnapshotCapture` + `MockRestoreAS` for already-running
  L2 leaves. FileJSON/stdout/stderr/AS-script asserts.

## Decision Tree

```
sessions-save/
├── help/
│   ├── show-usage/              sessions -h mentions save + restore
│   ├── color-flags/             save -h / restore -h mention --color + --no-color
│   └── ignore-macos-space/      save -h / restore -h mention --ignore-macos-space
├── save/
│   ├── dry-run/                 plan only; no file (auto color; pipe OK monochrome)
│   ├── write/                   writes version + saved_at; no restored_at
│   ├── zero/                    0 critical; no write
│   ├── pending-non-tty/         existing pending + non-TTY → error
│   ├── color/
│   │   ├── force-on/            save --dry-run --color → ANSI (green/bold)
│   │   ├── force-off/           --no-color → no ESC
│   │   └── conflict/            --color --no-color → non-zero + together msg
│   ├── stream/
│   │   └── order/               two critical windows; W1 before last ListTabs
│   └── space/                   ★ P2 macOS Space on save
│       ├── dry-run-label/       plan shows space N (Desktop N+1); not iterm_window_id
│       ├── write-fields/        FileJSON emits "space" (incl 0); not ignore
│       ├── ignore-omit/         --ignore-macos-space → no space / iterm_window_id keys
│       └── resolve-fail/        resolve fail → space 0 + stderr warning
└── restore/
    ├── dry-run/                 plan shows cd + resume; not stamped
    ├── consumed/                restored_at set → error
    ├── color/
    │   └── force-on/            restore --dry-run --color → ANSI
    ├── space/                   ★ P2 macOS Space on restore
    │   ├── dry-run-recorded/    seed space 2 → space 2 (Desktop 3); not stamped
    │   ├── dry-run-clamp/       seed space 16 → plans space 0 (Desktop 1) + warn
    │   ├── dry-run-missing/     no space key → space 0 (Desktop 1)
    │   └── dry-run-ignore/      --ignore-macos-space → no space lines; not stamped
    └── already-running/         ★ skip tabs already live (kind+session_id / mark msg)
        ├── dry-run-agent-hit/   grok session_id match → warn + skip; mark would restore
        ├── dry-run-mark-hit/    mark message match → warn + skip
        ├── dry-run-miss/        live unrelated → no already-running warn; all would restore
        ├── dry-run-mixed-window/ same window hit+miss; header remaining counts; both actions
        ├── dry-run-all-skip/    0 would restore; all skip lines; not stamped
        ├── dry-run-codex/       codex key match (not confused with grok)
        ├── live-partial-skip/   warn; AS only remaining; restored_at; summary skip count
        ├── live-all-skip/       stamp; 0 restored; no create-window AS
        └── scan-fail/           capture soft-fail → full plan / no skip; dry-run OK
```

## Test Index

| Leaf | Description | Expect |
|------|-------------|--------|
| `help/show-usage/` | Help mentions `save`, `restore`, `snapshot` | GREEN |
| `help/color-flags/` | save + restore help document `--color` / `--no-color` | RED until flags documented |
| `help/ignore-macos-space/` | save + restore help document `--ignore-macos-space` | RED until flag documented |
| `save/dry-run/` | Would save + critical ids; no file | GREEN (monochrome OK) |
| `save/write/` | file version, saved_at, kinds | GREEN |
| `save/zero/` | 0 critical; no write | GREEN |
| `save/pending-non-tty/` | pending + non-TTY error | GREEN |
| `save/color/force-on/` | `--color` ANSI on plan | RED |
| `save/color/force-off/` | `--no-color` no `\x1b` | RED until flag accepted (then GREEN if monochrome default) |
| `save/color/conflict/` | both flags → cannot be specified together | RED |
| `save/stream/order/` | W1 before last ListTabs | RED |
| `save/space/dry-run-label/` | dry-run plan includes `space N (Desktop N+1)`; no `iterm_window_id` text | RED |
| `save/space/write-fields/` | checkpoint JSON always has `"space"` when not ignore | RED |
| `save/space/ignore-omit/` | `--ignore-macos-space` omits `space` and `iterm_window_id` keys | RED (unknown flag today) |
| `save/space/resolve-fail/` | resolve fail → `"space": 0` + stderr warning | RED |
| `restore/dry-run/` | Would restore; not stamped | GREEN |
| `restore/consumed/` | consumed error | GREEN |
| `restore/color/force-on/` | restore `--color` ANSI | RED |
| `restore/space/dry-run-recorded/` | seed space 2 → plan `space 2 (Desktop 3)`; not stamped | RED |
| `restore/space/dry-run-clamp/` | space≥16 → plan `space 0 (Desktop 1)` + warn; not stamped | RED |
| `restore/space/dry-run-missing/` | missing space → plan `space 0 (Desktop 1)` | RED |
| `restore/space/dry-run-ignore/` | ignore → no space placement lines; not stamped | RED (unknown flag today) |
| `restore/already-running/dry-run-agent-hit/` | live grok session_id → warn+skip; remaining would restore; not stamped | RED |
| `restore/already-running/dry-run-mark-hit/` | live mark message → warn+skip; not stamped | RED |
| `restore/already-running/dry-run-miss/` | live unrelated critical → no already-running warn; all would restore | RED* |
| `restore/already-running/dry-run-mixed-window/` | same window hit+miss; reduced header counts; skip + would-restore lines | RED |
| `restore/already-running/dry-run-all-skip/` | all skip; 0 would create; not stamped | RED |
| `restore/already-running/dry-run-codex/` | codex kind+session_id hit → warn+skip | RED |
| `restore/already-running/live-partial-skip/` | live skip one; AS only remaining; stamp + skip summary | RED |
| `restore/already-running/live-all-skip/` | all skip; stamp; no create-window AS | RED |
| `restore/already-running/scan-fail/` | capture fail soft-warn; full plan no skip | RED |

\* miss may look GREEN if product never warns without a hit; asserts require
explicit would-restore action markers / no skip meta once format lands — RED until
tab-level action lines exist.

## How to Run

```sh
doctest vet ./tests/iterm2/sessions-save
doctest test ./tests/iterm2/sessions-save
# Space leaves only:
doctest test ./tests/iterm2/sessions-save/help/ignore-macos-space
doctest test ./tests/iterm2/sessions-save/save/space
doctest test ./tests/iterm2/sessions-save/restore/space
# Already-running skip leaves:
doctest test ./tests/iterm2/sessions-save/restore/already-running
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
	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

const (
	ModeHelp    = "help"
	ModeSave    = "save"
	ModeRestore = "restore"
)

const (
	fixtureGrokSessionID  = "019fabcdef-1234-5678-9abc-def012345678"
	fixtureCodexSessionID = "codex-019f-aaaa-bbbb-cccc-ddddeeeeffff"
	fixtureMarkMessage    = "still waiting for CI"
	fixtureLiveWindowID   = uint64(9001)
)

// RestoreLiveFixture presets for already-running leaves (ModeRestore).
const (
	RestoreLiveAgentHit = "agent-hit"    // live grok matches fixtureGrokSessionID
	RestoreLiveMarkHit  = "mark-hit"     // live mark matches fixtureMarkMessage
	RestoreLiveMiss     = "miss"         // live unrelated grok+mark
	RestoreLiveMixed    = "mixed"        // live only grok (same window hit+miss seed)
	RestoreLiveAllSkip  = "all-skip"     // live grok+mark both match seed
	RestoreLiveCodexHit = "codex-hit"    // live codex matches fixtureCodexSessionID
	RestoreLivePartial  = "partial-skip" // live only grok (live AS path)
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

	// IgnoreMacOSSpace maps to --ignore-macos-space on save|restore.
	// Zero value false keeps existing leaves unchanged.
	IgnoreMacOSSpace bool

	// Install critical fixture (one window: grok + mark + idle).
	UseCriticalFixture bool
	// Idle-only fixture (zero critical).
	UseIdleOnlyFixture bool
	// Two windows each with ≥1 critical tab (stream order).
	UseTwoCriticalWindows bool

	// ObserveStreamOrder records SawW1BeforeLastListTabs via OnListTabs probe.
	ObserveStreamOrder bool

	// Pre-seed checkpoint at FilePath before run (struct path).
	SeedDoc *iterm2.SaveDocument
	// SeedRawJSON if non-empty, written to FilePath instead of SeedDoc.
	// Prefer for restore Space seeds that include "space" / "iterm_window_id"
	// without requiring production SaveWindow fields at design time.
	SeedRawJSON string

	// RestoreLiveFixture installs a live phased snapshot on ModeRestore so
	// already-running scan can match checkpoint tabs. Empty = no live fixture
	// (existing restore leaves unchanged).
	RestoreLiveFixture string
	// FailSnapshotCapture installs a collector that fails capture (iTerm not
	// running) — for scan-fail soft-warn leaves.
	FailSnapshotCapture bool
	// MockRestoreAS mocks sessionsRunRestoreAS and records scripts in Response.
	// Also installs a Space MockBackend so live placement does not hit Mission Control.
	MockRestoreAS bool

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

	// Restore AS capture when MockRestoreAS is set.
	RestoreASScripts   []string
	RestoreASCallCount int
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

// installRestoreFailCapture makes CaptureSnapshot fail (iTerm not running).
func installRestoreFailCapture(t *testing.T) {
	t.Helper()
	iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
		ITermRunning: false,
		Windows:      nil,
		Hostname:     "testhost",
	})
}

// installRestoreLiveFixture installs live panes for already-running match tests.
// Also injects a fixed Space index resolver so warnings can mention space N.
func installRestoreLiveFixture(t *testing.T, kind string) {
	t.Helper()
	iterm2.SetSpaceIndexForWindowForTest(func(windowID uint64) (int, error) {
		if windowID == fixtureLiveWindowID {
			return 2, nil // space 2 (Desktop 3)
		}
		return 0, nil
	})
	t.Cleanup(func() { iterm2.SetSpaceIndexForWindowForTest(nil) })

	switch kind {
	case RestoreLiveAgentHit, RestoreLiveMixed, RestoreLivePartial:
		// One live grok matching fixtureGrokSessionID.
		iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
			ITermRunning: true,
			Windows: []iterm2.SnapshotWindow{{
				Index: 1, Name: "Live-Win", WindowID: fixtureLiveWindowID,
				Tabs: []iterm2.SnapshotTab{{
					Index: 1, Name: "live-grok",
					Sessions: []iterm2.SnapshotSession{{
						Index: 1, ID: "LIVE0001-0000-0000-0000-000000000001",
						TTY: "/dev/ttys010", Profile: "Default", Name: "grok-pane",
					}},
				}},
			}},
			BusyTTYs: []string{"ttys010"},
			CwdByTTY: map[string]string{"ttys010": "/proj/live-grok"},
			AgentResolveByTTY: map[string]iterm2.AgentResolveFixture{
				"ttys010": {Kind: "grok", SessionID: fixtureGrokSessionID, Title: "live-grok"},
			},
			Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
			Hostname: "testhost",
		})
	case RestoreLiveMarkHit:
		iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
			ITermRunning: true,
			Windows: []iterm2.SnapshotWindow{{
				Index: 1, Name: "Live-Win", WindowID: fixtureLiveWindowID,
				Tabs: []iterm2.SnapshotTab{{
					Index: 1, Name: "live-mark",
					Sessions: []iterm2.SnapshotSession{{
						Index: 1, ID: "LIVE0002-0000-0000-0000-000000000002",
						TTY: "/dev/ttys011", Profile: "Default", Name: "mark-pane",
					}},
				}},
			}},
			BusyTTYs: []string{"ttys011"},
			BusyLeafByTTY: map[string]string{
				"ttys011": "mark " + fixtureMarkMessage,
			},
			CwdByTTY: map[string]string{"ttys011": "/proj/live-mark"},
			Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
			Hostname: "testhost",
		})
	case RestoreLiveMiss:
		// Live critical panes with different ids/messages than the seed.
		iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
			ITermRunning: true,
			Windows: []iterm2.SnapshotWindow{{
				Index: 1, Name: "Live-Unrelated", WindowID: fixtureLiveWindowID,
				Tabs: []iterm2.SnapshotTab{
					{Index: 1, Name: "other-grok", Sessions: []iterm2.SnapshotSession{{
						Index: 1, ID: "LIVE0003-0000-0000-0000-000000000003",
						TTY: "/dev/ttys020", Profile: "Default", Name: "other-grok",
					}}},
					{Index: 2, Name: "other-mark", Sessions: []iterm2.SnapshotSession{{
						Index: 1, ID: "LIVE0004-0000-0000-0000-000000000004",
						TTY: "/dev/ttys021", Profile: "Default", Name: "other-mark",
					}}},
				},
			}},
			BusyTTYs: []string{"ttys020", "ttys021"},
			BusyLeafByTTY: map[string]string{
				"ttys021": "mark unrelated work",
			},
			CwdByTTY: map[string]string{
				"ttys020": "/proj/other-grok",
				"ttys021": "/proj/other-mark",
			},
			AgentResolveByTTY: map[string]iterm2.AgentResolveFixture{
				"ttys020": {Kind: "grok", SessionID: "other-grok-session-id-zzzz", Title: "other"},
			},
			Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
			Hostname: "testhost",
		})
	case RestoreLiveAllSkip:
		// Both grok + mark matching the standard two-tab seed.
		iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
			ITermRunning: true,
			Windows: []iterm2.SnapshotWindow{{
				Index: 1, Name: "Live-Both", WindowID: fixtureLiveWindowID,
				Tabs: []iterm2.SnapshotTab{
					{Index: 1, Name: "live-grok", Sessions: []iterm2.SnapshotSession{{
						Index: 1, ID: "LIVE0005-0000-0000-0000-000000000005",
						TTY: "/dev/ttys030", Profile: "Default", Name: "grok-pane",
					}}},
					{Index: 2, Name: "live-mark", Sessions: []iterm2.SnapshotSession{{
						Index: 1, ID: "LIVE0006-0000-0000-0000-000000000006",
						TTY: "/dev/ttys031", Profile: "Default", Name: "mark-pane",
					}}},
				},
			}},
			BusyTTYs: []string{"ttys030", "ttys031"},
			BusyLeafByTTY: map[string]string{
				"ttys031": "mark " + fixtureMarkMessage,
			},
			CwdByTTY: map[string]string{
				"ttys030": "/proj/live-grok",
				"ttys031": "/proj/live-mark",
			},
			AgentResolveByTTY: map[string]iterm2.AgentResolveFixture{
				"ttys030": {Kind: "grok", SessionID: fixtureGrokSessionID, Title: "live-grok"},
			},
			Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
			Hostname: "testhost",
		})
	case RestoreLiveCodexHit:
		iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
			ITermRunning: true,
			Windows: []iterm2.SnapshotWindow{{
				Index: 1, Name: "Live-Codex", WindowID: fixtureLiveWindowID,
				Tabs: []iterm2.SnapshotTab{{
					Index: 1, Name: "live-codex",
					Sessions: []iterm2.SnapshotSession{{
						Index: 1, ID: "LIVE0007-0000-0000-0000-000000000007",
						TTY: "/dev/ttys040", Profile: "Default", Name: "codex-pane",
					}},
				}},
			}},
			BusyTTYs: []string{"ttys040"},
			CwdByTTY: map[string]string{"ttys040": "/proj/live-codex"},
			AgentResolveByTTY: map[string]iterm2.AgentResolveFixture{
				"ttys040": {Kind: "codex", SessionID: fixtureCodexSessionID, Title: "live-codex"},
			},
			Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
			Hostname: "testhost",
		})
	default:
		t.Fatalf("unknown RestoreLiveFixture %q", kind)
	}
}

// seedGrokMarkDoc is the standard two-tab checkpoint for already-running leaves.
func seedGrokMarkDoc() *iterm2.SaveDocument {
	return &iterm2.SaveDocument{
		Version: 1,
		SavedAt: "2026-07-25T18:00:00+0800",
		Host:    "testhost",
		Source:  "kool-iterm2-sessions-save",
		Summary: iterm2.SaveSummary{Windows: 1, Tabs: 2, Sessions: 2, ByKind: map[string]int{"grok": 1, "mark": 1}},
		Windows: []iterm2.SaveWindow{{
			SourceIndex: 1,
			Name:        "Win-Crit",
			Tabs: []iterm2.SaveTab{
				{
					Name: "grok-tab", Cwd: "/proj/a", Kind: "grok",
					SessionID: fixtureGrokSessionID,
					ResumeCmd: "grok --resume " + fixtureGrokSessionID,
				},
				{
					Name: "mark-tab", Cwd: "/proj/b", Kind: "mark",
					Message:   fixtureMarkMessage,
					ResumeCmd: "mark '" + fixtureMarkMessage + "'",
				},
			},
		}},
	}
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

func appendSpaceFlags(args []string, req *Request) []string {
	if req.IgnoreMacOSSpace {
		args = append(args, "--ignore-macos-space")
	}
	return args
}

// seedCheckpoint writes SeedRawJSON (preferred) or SeedDoc to path.
func seedCheckpoint(path string, req *Request) error {
	if req.SeedRawJSON != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(req.SeedRawJSON), 0o644)
	}
	if req.SeedDoc != nil {
		return iterm2.WriteSaveDocument(path, req.SeedDoc)
	}
	return nil
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	path := resolveFile(req)

	if err := seedCheckpoint(path, req); err != nil {
		return nil, err
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
		args = appendSpaceFlags(args, req)
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

		// Live snapshot for already-running scan (optional; zero-value = prior leaves).
		if req.FailSnapshotCapture {
			installRestoreFailCapture(t)
		} else if req.RestoreLiveFixture != "" {
			installRestoreLiveFixture(t, req.RestoreLiveFixture)
		} else if req.UseCriticalFixture {
			installCritical(t, nil)
		} else if req.UseIdleOnlyFixture {
			installIdleOnly(t)
		} else if req.UseTwoCriticalWindows {
			installTwoCritical(t, nil)
		}

		var asScripts []string
		if req.MockRestoreAS {
			// Avoid live Mission Control / AppleScript on live restore leaves.
			iterm2.SetSpaceBackendForTest(&space.MockBackend{Desktops: []int{1, 2, 3}})
			t.Cleanup(func() { iterm2.SetSpaceBackendForTest(nil) })
			iterm2.SetSessionsRunRestoreASForTest(func(script string) (string, error) {
				asScripts = append(asScripts, script)
				return "", nil
			})
			t.Cleanup(func() { iterm2.SetSessionsRunRestoreASForTest(nil) })
		}

		args := []string{"sessions", "restore", "--file", path}
		if req.DryRun {
			args = append(args, "--dry-run")
		}
		args = appendColorFlags(args, req)
		args = appendSpaceFlags(args, req)
		args = append(args, req.ExtraArgs...)
		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		resp := &Response{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: code,
		}
		if req.MockRestoreAS {
			resp.RestoreASScripts = append([]string(nil), asScripts...)
			resp.RestoreASCallCount = len(asScripts)
		}
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
