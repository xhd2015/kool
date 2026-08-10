# kool iterm2 sessions save / restore

Checkpoint critical **grok**, **codex**, and **mark** tabs to
`~/.config/iterm2/sessions-save.json` (override with `--file`), then restore
windows with `cd` + resume commands. **P2** also records macOS Space
(`space` + `iterm_window_id`) and restores placement by Space index.
**Multi-app** records per-window canonical **`app`** (home + system iTerm
installs) on save. Restore **default** picks one global path target
(prefer `~/Applications/iTerm.app` when that install exists on disk); opt-in
**`--same-app`** recreates each window in its recorded `app`.

**Sibling of** `./tests/iterm2/sessions` (snapshot) and `sessions-p4` (enrich).

## Version

0.0.4

**Classic TDD (color + streaming + Space + already-running + multi-app +
restore app-target):** dry-run color flags, progressive save stream, **macOS
Space**, **restore already-running skip**, **multi-app / `app` field**, and
**restore prefer-home / `--same-app`** leaves are intentionally **RED** until
the implementer lands those behaviors. Existing dry-run / write / restore /
help leaves stay contracted and should remain **GREEN** on monochrome
buffered output (zero-value Request defaults — no Space / already-running /
multi-app / RestoreAppDisk fixture flags). After already-running lands,
restore always attempts a live capture; capture fail is soft so prior leaves
without fixtures stay GREEN. Extra dry-run meta lines (`restore target`,
`recorded app`) are additive: prior leaves that only assert Would restore /
resume cmds stay GREEN if product always prints them.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — invokes `kool iterm2 sessions save|restore …`.
- **kool CLI / handler** — `tools/iterm2` routes `sessions save` / `sessions restore`;
  flags `--dry-run`, `--file`, **`--color`**, **`--no-color`**,
  **`--ignore-macos-space`**, save-only **`--spaces LIST`**, restore-only
  **`--same-app`**, `-h/--help`.
- **SnapshotCollector** — injectable via `InstallPhasedFixtureCollectorForTest`;
  phased `ListWindows` / `ListTabsAndSessions` + enrich + agent resolve.
  Fixtures may carry per-window **WindowID** (iTerm/CG window number) and
  (implementer) **App** for multi-app dual-source tags.
- **App preflight** — save-only: resolve bare `tell application "iTerm2"` →
  AS window ids → CG owner PID majority → `proc_pidpath` → canonical **asApp**;
  discover running home + `/Applications` iTerm binaries; build source list
  (bare tagged asApp ∪ path-tells for other running apps); merge windows;
  hard-dedupe by `iterm_window_id`; renumber `source_index` 1…N. Dual collapse
  (other path yields no new ids) → stderr **warning**, partial save, exit 0.
- **Space index resolver** — P1 `space.SpaceIndexForWindow(windowID)` (via go.mod
  replace to `external/dot-pkgs-…/go-pkgs`). Save maps window id → 0-based
  Desktop index. Injectable for tests (implementer hook; no live WindowServer).
  **`--spaces`** runs **after** multi-app merge.
- **Space Backend** — Create / Switch / Highest for restore placement
  (`MockBackend` pattern from `computer-use/macos/space`). Live restore only;
  dry-run never calls it.
- **Critical filter** — keeps grok/codex (session_id) and live mark panes;
  builds `SaveDocument`.
- **Save plan streamer** — on save `--dry-run`, after each window is captured
  and classified with ≥1 critical tab, writes that **window block** to stdout
  immediately (includes **space N (Desktop N+1)** meta when Space recording is
  on; gray **`app  <canonical>`** meta when `app` non-empty); after all
  windows, writes the **footer**. No “scanning…” header. First window has no
  leading blank line.
- **Restore app resolver** — chooses create/tell path by **disk presence**
  (injectable for tests), not by which process is running:
  - **Default** (no `--same-app`): one global target for all windows —
    prefer `~/Applications/iTerm.app` when that path exists; else system;
    else warn + bare `"iTerm2"`. Does **not** use per-window recorded `app`
    for create/tell.
  - **`--same-app`**: per-window path from checkpoint `windows[].app`
    (canonical home/system); empty/missing → same prefer-home fallback + warn.
  - Missing target on disk → warn + fallback (home → system → bare).
- **Restore planner** — restore `--dry-run` prints header → optional global
  **`restore target  <path>`** → each window (space meta; under default,
  **`recorded app  <path>`** only when it differs from restore target; under
  `--same-app`, per-window **`app  <path>`** create target) / tab → footer.
  No Switch/Create/AS side effects for placement on dry-run.
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
  probe; multi-app topology fixtures (`UseMultiApp*`); `SeedRawJSON` for
  checkpoints with `space` / `iterm_window_id` / `app` without requiring
  production struct fields at design time. No live iTerm / Mission Control /
  dual-process.

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
| `--spaces LIST` on save | 0-based comma list; keep only windows whose resolved `space` is in LIST. Hard error with `--ignore-macos-space`. Soft resolve fail → `space=0` for membership. Runs **after** multi-app merge. |
| `filter.spaces` | When `--spaces` used: sorted unique list in checkpoint. Omitted when flag not set. Restore ignores for placement. |

### Checkpoint App field (`SaveWindow.app`)

| Field / rule | Behavior |
|--------------|----------|
| `app` | Optional JSON string. Only **`~/Applications/iTerm.app`** or **`/Applications/iTerm.app`**. Always set when known (D1). Version stays 1 (additive). |
| Home form | Always `~/…`, never expanded `/Users/…`. Non-standard home install (e.g. `.bak`) still records `~/Applications/iTerm.app` (optional one warning). |
| Multi-app save | Sources = bare AS tagged asApp ∪ path-tells for other running canonical installs; hard-dedupe by `iterm_window_id`; renumber `source_index` globally 1…N. |
| Dual collapse | Other path yields no new ids → stderr warning + partial save, exit 0 (D2). |
| Dry-run meta (save) | When `app` non-empty, window block includes gray **`app  <path>`** line (D4). |
| Restore default | One global **`restore target`** (prefer home when on disk). Recorded `app` is **not** used for create. Dry-run shows **`recorded app`** only when it differs from restore target (Open2). |
| Restore `--same-app` | Per-window create target = recorded `app` (canonical). Dry-run **`app  <path>`** per window. Empty app → prefer-home fallback + warn (R7). |
| Restore fallback | Neither install on disk → warn + bare `"iTerm2"` (Open1/R8). Path missing → warn + chain home→system→bare (R6). |
| Snapshot | Multi-app capture is **save-only** (`CaptureOpts` / save path); `sessions snapshot` unchanged (D3). |

### Implementer inject (tests — multi-app + restore app disk)

Expected product seams (parallel-safe; `t.Cleanup`; no `Setenv`/`Chdir`):

- `SnapshotWindow.App` (fixture tag) and/or preflight inject for **asApp** + **runningApps**
- Multi-app fixture collector: two sources of windows with different App tags
  (home vs system); merge + dedupe by `iterm_window_id` in product
- Doctest harness uses `UseMultiAppFixture` / `UseMultiAppDedupeFixture` /
  `UseMultiAppSpacesFixture` + `FixtureApp` (assert target for single-app).
  Until inject lands, topology fixtures still drive RED asserts on missing
  `"app"` / dual app values / collapse warning.
- **Restore disk presence** (required for deterministic L2 app-target leaves):
  land **`SetRestoreAppDiskForTest`** (name may match; mirror
  `SetMultiAppPreflightForTest` style — mutex + `t.Cleanup(nil)`):
  - Signature sketch: `func SetRestoreAppDiskForTest(homeExists, systemExists bool)`
    or a small struct / optional-pointer override of “which installs exist”.
  - When set, restore target resolution uses inject instead of `os.Stat` on
    `~/Applications/iTerm.app` and `/Applications/iTerm.app`.
  - Doctest `Request.RestoreAppDisk` values map in `Run`:
    - `"both"` → home+system true
    - `"system"` → only system
    - `"home"` → only home
    - `"neither"` → both false
  - Designer **does not** call the hook until it exists (would break compile of
    the whole tree). Implementer wires the call in root `Run` ModeRestore.

### Save (Space)

1. Snapshot captures per-window iTerm window id (fixture `WindowID` / AS `id of window`).
2. Unless ignore: resolve `space` via `SpaceIndexForWindow(iterm_window_id)` (injectable).
3. Resolve fail / non-type-0 → `space=0` + **warning** (stderr yellow); still may emit `iterm_window_id` if known.
4. Always emit `space` when not ignore (including 0).
5. When `--spaces`: after resolve, drop non-matching critical windows; recompute summary;
   set `filter.spaces`; if any dropped → stderr warning footer
   `skipped N windows not matching --spaces …`.

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
| Gray | path, cwd, resume_cmd lines, counts meta, space meta, **`restore target`**, **`recorded app`**, restore **`app`** (same-app), `(dry-run: not written\|not applied)`, saved_at meta |

Stderr: existing yellow `warning:`, red `Error:` (unchanged).

### Streaming order (save --dry-run, observable)

1. For each window after enrich: classify critical tabs
2. If ≥1 critical → **write that window block** (`W{n}` …) to stdout immediately.
   **First** critical window has **no** leading blank line; later windows are
   separated by one blank line.
3. After all windows → footer: green Would save + gray path + dry-run note

When `ListTabsAndSessions` for the **last** window begins and ≥2 windows have
critical content, stdout must already contain `W1` (stream probe).

### Inject (tests)

- `InstallPhasedFixtureCollectorForTest` + `AgentResolveByTTY` / `BusyLeafByTTY`
- **Implementer surface (expected):**
  - `SnapshotWindow.WindowID` (uint64) for fixture `iterm_window_id`
  - `SnapshotWindow.App` / `SaveWindow.App` (JSON `app`; canonical strings only)
  - multi-app / preflight inject (asApp, runningApps, dual-source collectors)
  - `SaveWindow.Space` / `SaveWindow.ItermWindowID` (JSON `space` / `iterm_window_id`)
  - test hook to inject Space index resolver (fixed index / error)
  - test hook to inject Space Backend + restore AppleScript for live placement
  - already-running scan after checkpoint load (dry-run + live); soft capture fail
  - **`SetRestoreAppDiskForTest`** for home/system install existence (restore
    prefer-home / `--same-app` fallback); see Implementer inject above
  - restore create/tell uses path `tell application "/…/iTerm.app"` (or expanded
    home path) when target resolved — not only bare `"iTerm2"`
- Doctest harness: `SeedRawJSON` / `SeedDoc`, `--ignore-macos-space` via Request,
  `RestoreLiveFixture` + `FailSnapshotCapture` + `MockRestoreAS` for already-running
  L2 leaves; `UseMultiApp*` + `FixtureApp` for multi-app L2 leaves;
  **`SameApp`** + **`RestoreAppDisk`** for restore app-target leaves (implementer
  wires disk inject in `Run`). FileJSON/stdout/stderr/AS-script asserts.

## Decision Tree

```
sessions-save/
├── help/
│   ├── show-usage/              sessions -h mentions save + restore
│   ├── color-flags/             save -h / restore -h mention --color + --no-color
│   ├── ignore-macos-space/      save -h / restore -h mention --ignore-macos-space
│   ├── spaces-flag/             save -h mentions --spaces
│   ├── multi-app/               save -h mentions dual installs / app / preferred restore
│   └── same-app/                ★ restore -h: --same-app + prefer home when multi
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
│   ├── space/                   ★ P2 macOS Space on save
│   │   ├── dry-run-label/       plan shows space N (Desktop N+1); not iterm_window_id
│   │   ├── write-fields/        FileJSON emits "space" (incl 0); not ignore
│   │   ├── ignore-omit/         --ignore-macos-space → no space / iterm_window_id keys
│   │   ├── resolve-fail/        resolve fail → space 0 + stderr warning
│   │   ├── filter-keep/         --spaces 0 keeps space-0 window; skip warn + filter.spaces
│   │   ├── filter-drop-all/     --spaces 1 drops all; 0 critical + skip warn; no write
│   │   └── conflict-ignore/     --spaces + --ignore-macos-space → error
│   └── app/                     ★ multi-app + SaveWindow.app (save only)
│       ├── single-write/        single system fixture → FileJSON "app" system path
│       ├── dual-merge/          both canonical apps; counts; no dup iterm_window_id
│       ├── dry-run-meta/        dry-run gray app meta line; no leading blank; exit 0
│       ├── dedupe-collapse/     dual source same ids → no doubles + warn; exit 0
│       └── filter-spaces/       --spaces after merge; kept windows retain correct app
└── restore/
    ├── dry-run/                 plan shows cd + resume; not stamped (no app seed)
    ├── consumed/                restored_at set → error
    ├── app-ignored/             ★ rewritten: default ignores app for create; may show
    │                            restore target + recorded app when differs
    ├── app-target/              ★ prefer-home + --same-app path targeting
    │   ├── default-prefer-home/ both installs; target home; recorded system differs
    │   ├── default-only-system/ only system on disk → target system
    │   ├── default-only-home/   only home on disk → target home
    │   ├── same-app-system/     --same-app; plan app system (+ live path-tell AS)
    │   ├── same-app-home/       --same-app; plan app home
    │   ├── same-app-mixed/      --same-app; two windows different apps
    │   ├── same-app-empty-fallback/ --same-app; empty app → warn + prefer-home
    │   └── neither-install-bare/ neither on disk → warn + bare iTerm2 plan
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
| `help/spaces-flag/` | save help documents `--spaces` | GREEN |
| `help/multi-app/` | save help mentions dual installs / app / preferred restore | RED until help documents multi-app |
| `help/same-app/` | restore help documents `--same-app` + prefer home when multiple | RED until flag/help lands |
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
| `save/space/filter-keep/` | `--spaces 0` keeps space-0 window; filter.spaces + skip warn | RED |
| `save/space/filter-drop-all/` | `--spaces 1` drops all; no write | RED |
| `save/space/conflict-ignore/` | `--spaces` + `--ignore-macos-space` → error | RED |
| `save/app/single-write/` | FileJSON windows include `"app": "/Applications/iTerm.app"` | RED until app field |
| `save/app/dual-merge/` | both canonical apps; critical counts; no duplicate window ids | RED until multi-app merge |
| `save/app/dry-run-meta/` | stdout `app  …` meta; no leading blank; exit 0; no file | RED until dry-run app meta |
| `save/app/dedupe-collapse/` | dual same ids → one window + stderr warn; exit 0 | RED until multi-app dedupe |
| `save/app/filter-spaces/` | `--spaces` after multi-app; kept window retains correct `app` | RED until multi-app + spaces |
| `restore/dry-run/` | Would restore; not stamped (seed without app) | GREEN |
| `restore/app-ignored/` | default: seed system app + both installs; `restore target` home; `recorded app` system; no same-app create `app` line; not stamped | RED (rewritten contract) |
| `restore/app-target/default-prefer-home/` | both installs; recorded system → target home + recorded app line | RED |
| `restore/app-target/default-only-system/` | only system on disk → restore target system | RED |
| `restore/app-target/default-only-home/` | only home on disk → restore target home | RED |
| `restore/app-target/same-app-system/` | `--same-app`; plan/AS targets system path | RED |
| `restore/app-target/same-app-home/` | `--same-app`; plan targets home | RED |
| `restore/app-target/same-app-mixed/` | `--same-app`; two windows different apps | RED |
| `restore/app-target/same-app-empty-fallback/` | `--same-app`; empty app → warn + prefer-home | RED |
| `restore/app-target/neither-install-bare/` | neither install → warn + bare iTerm2 plan | RED |
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
# Multi-app / app field leaves:
doctest test ./tests/iterm2/sessions-save/save/app
doctest test ./tests/iterm2/sessions-save/restore/app-ignored
doctest test ./tests/iterm2/sessions-save/help/multi-app
# Restore prefer-home / --same-app:
doctest test ./tests/iterm2/sessions-save/help/same-app
doctest test ./tests/iterm2/sessions-save/restore/app-target
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

	// Spaces maps to save --spaces LIST (empty = omit flag).
	Spaces string

	// Install critical fixture (one window: grok + mark + idle).
	UseCriticalFixture bool
	// Idle-only fixture (zero critical).
	UseIdleOnlyFixture bool
	// Two windows each with ≥1 critical tab (stream order).
	UseTwoCriticalWindows bool
	// Two critical windows with WindowIDs resolving to space 0 and 2.
	UseTwoCriticalSpacesFixture bool

	// --- Multi-app / SaveWindow.app (save path). Zero-value = legacy single-app. ---
	// FixtureApp is the expected canonical app for single-source leaves (assert
	// target). Product must set app when known (D1). Typical:
	//   "/Applications/iTerm.app" or "~/Applications/iTerm.app"
	FixtureApp string
	// UseMultiAppFixture: dual-source merge topology (system + home apps,
	// distinct iterm_window_ids). Implementer tags per-window App on merge.
	UseMultiAppFixture bool
	// UseMultiAppDedupeFixture: dual-source with identical iterm_window_ids
	// (collapse → one window + stderr warning, exit 0).
	UseMultiAppDedupeFixture bool
	// UseMultiAppSpacesFixture: dual apps with FixedSpace 0 (system) + 2 (home)
	// for --spaces-after-merge leaves.
	UseMultiAppSpacesFixture bool

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

	// --- Restore app-target / --same-app (zero-value = prior leaves unchanged). ---
	// SameApp maps to restore --same-app (opt-in per-window recorded app create).
	// Default false = one global prefer-home target; recorded app not used for create.
	SameApp bool
	// RestoreAppDisk injects which canonical installs exist on disk for restore
	// target resolution (parallel-safe via implementer SetRestoreAppDiskForTest).
	// Empty = no inject (product live disk check). Values:
	//   RestoreDiskBoth | RestoreDiskSystem | RestoreDiskHome | RestoreDiskNeither
	// Designer stores the field; implementer wires the product hook in Run
	// (calling a missing symbol would break compile of the whole tree).
	RestoreAppDisk string

	// Force non-TTY for overwrite checks (default true for save tests).
	NonTTY *bool

	WorkingDir string
}

// RestoreAppDisk values for Request.RestoreAppDisk (disk presence inject).
const (
	RestoreDiskBoth    = "both"    // home + system exist
	RestoreDiskSystem  = "system"  // only /Applications/iTerm.app
	RestoreDiskHome    = "home"    // only ~/Applications/iTerm.app
	RestoreDiskNeither = "neither" // neither install on disk
)

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

const (
	fixtureSpaceWindowID0 = uint64(1001)
	fixtureSpaceWindowID2 = uint64(1002)
)

// Canonical app paths (D9). Home form always ~/… never /Users/….
const (
	fixtureAppSystem = "/Applications/iTerm.app"
	fixtureAppHome   = "~/Applications/iTerm.app"
)

const (
	fixtureMultiWindowIDSystem = uint64(5001)
	fixtureMultiWindowIDHome   = uint64(5002)
	fixtureMultiWindowIDShared = uint64(5003) // dedupe: same id from both sources
)

// installTwoCriticalSpaces installs W1@space0 and W2@space2 via FixedSpace
// (parallel-safe; no global SpaceIndex resolver).
func installTwoCriticalSpaces(t *testing.T) {
	t.Helper()
	space0, space2 := 0, 2
	wins := twoCriticalWindows()
	wins[0].WindowID = fixtureSpaceWindowID0
	wins[0].Name = "On-Space-0"
	wins[0].FixedSpace = &space0
	wins[1].WindowID = fixtureSpaceWindowID2
	wins[1].Name = "On-Space-2"
	wins[1].FixedSpace = &space2
	iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
		ITermRunning: true,
		Windows:      wins,
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
	})
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

// installMultiAppMerge installs the dual-source merge topology: W1 system-app
// window + W2 home-app window (distinct WindowIDs, both critical).
//
// Until implementer lands MultiApp capture + SnapshotWindow.App / preflight
// tags, save emits windows without dual `"app"` values → dual-merge RED.
// When App field lands, implementer should tag:
//   wins[0].App = fixtureAppSystem; wins[1].App = fixtureAppHome
// (or inject dual-source collectors that set App per source).
func installMultiAppMerge(t *testing.T) {
	t.Helper()
	wins := twoCriticalWindows()
	wins[0].WindowID = fixtureMultiWindowIDSystem
	wins[0].Name = "From-System"
	wins[1].WindowID = fixtureMultiWindowIDHome
	wins[1].Name = "From-Home"
	iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
		ITermRunning: true,
		Windows:      wins,
		BusyTTYs:     []string{"ttys001", "ttys002"},
		BusyLeafByTTY: map[string]string{
			"ttys002": "mark still waiting for CI",
		},
		CwdByTTY: map[string]string{
			"ttys001": "/proj/system-grok",
			"ttys002": "/proj/home-mark",
		},
		AgentResolveByTTY: map[string]iterm2.AgentResolveFixture{
			"ttys001": {Kind: "grok", SessionID: fixtureGrokSessionID, Title: "system-grok"},
		},
		Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname: "testhost",
	})
}

// installMultiAppDedupe installs a single critical window with a shared
// WindowID. Dual-running second source would yield the same iterm_window_id
// (collapse → no doubles + stderr warning, exit 0). Without multi-app inject,
// product does not warn → dedupe-collapse RED.
func installMultiAppDedupe(t *testing.T) {
	t.Helper()
	wins := criticalWindows()
	wins[0].WindowID = fixtureMultiWindowIDShared
	wins[0].Name = "Shared-ID"
	iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
		ITermRunning: true,
		Windows:      wins,
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
	})
}

// installMultiAppSpaces installs dual-app topology with FixedSpace 0 (system)
// and 2 (home) for --spaces-after-merge leaves.
func installMultiAppSpaces(t *testing.T) {
	t.Helper()
	space0, space2 := 0, 2
	wins := twoCriticalWindows()
	wins[0].WindowID = fixtureMultiWindowIDSystem
	wins[0].Name = "System-Space-0"
	wins[0].FixedSpace = &space0
	wins[1].WindowID = fixtureMultiWindowIDHome
	wins[1].Name = "Home-Space-2"
	wins[1].FixedSpace = &space2
	iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
		ITermRunning: true,
		Windows:      wins,
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
	})
}

// fileJSONHasApp reports whether checkpoint JSON has "app": "<canonical>".
func fileJSONHasApp(fileJSON, app string) bool {
	// encoding/json emits a space after colon.
	if strings.Contains(fileJSON, `"app": "`+app+`"`) {
		return true
	}
	return strings.Contains(fileJSON, `"app":"`+app+`"`)
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
	if req.Spaces != "" {
		args = append(args, "--spaces", req.Spaces)
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
		} else if req.UseMultiAppDedupeFixture {
			installMultiAppDedupe(t)
		} else if req.UseMultiAppSpacesFixture {
			installMultiAppSpaces(t)
		} else if req.UseMultiAppFixture {
			installMultiAppMerge(t)
		} else if req.UseTwoCriticalSpacesFixture {
			installTwoCriticalSpaces(t)
		} else if req.UseTwoCriticalWindows {
			installTwoCritical(t, probe)
		} else if req.UseCriticalFixture {
			// Single-app path: UseCriticalFixture (FixtureApp is assert-only until
			// implementer tags SnapshotWindow.App / preflight asApp).
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
		// App-target leaves (RestoreAppDisk set) install idle-only when no other
		// live fixture is requested so SeedRawJSON session ids do not match a
		// dirty global/live collector (already-running pollution).
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
		} else if req.RestoreAppDisk != "" {
			// Prefer-home / --same-app leaves: isolate capture from live iTerm.
			installIdleOnly(t)
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

		// Disk inject for restore prefer-home / --same-app (exclusive hold until Cleanup).
		diskInjected := false
		switch req.RestoreAppDisk {
		case RestoreDiskBoth:
			iterm2.SetRestoreAppDiskForTest(true, true)
			diskInjected = true
		case RestoreDiskSystem:
			iterm2.SetRestoreAppDiskForTest(false, true)
			diskInjected = true
		case RestoreDiskHome:
			iterm2.SetRestoreAppDiskForTest(true, false)
			diskInjected = true
		case RestoreDiskNeither:
			iterm2.SetRestoreAppDiskForTest(false, false)
			diskInjected = true
		}
		if diskInjected {
			t.Cleanup(func() { iterm2.ClearRestoreAppDiskForTest() })
		}

		args := []string{"sessions", "restore", "--file", path}
		if req.DryRun {
			args = append(args, "--dry-run")
		}
		if req.SameApp {
			args = append(args, "--same-app")
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
