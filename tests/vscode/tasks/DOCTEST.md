# kool vscode tasks (P1 + P2 run backends + live iTerm2)

`kool vscode tasks` discovers a workspace `.vscode/tasks.json` (JSONC), lists /
finds / shows tasks, prints a **dry-run** execution plan, and **executes** a
resolved run plan via **local** process exec or **iterm2** multi-tab backends
(`auto` heuristic selects). Live iterm2 uses `lib.RunTabSet` with an optional
**test-only mock env seam** so CI never needs a real iTerm.

## Version

0.0.2

**Classic TDD (P2 + live iTerm2):** list/find/show/dry-run/match/help/validation
leaves from P1 stay valid. P2 adds real `run` backends and **retires** the P1
leaf `run/requires-dry-run`. Live iterm2 leaves under `run/backend/iterm2-*`
pin `RunTabSet` wiring via mock env (RED until implementer replaces the stub
`warning: live iterm2 run not configured; executing leaves locally`).

## DSN (Domain Specific Notion)

### Participants

- **kool CLI** — top-level `vscode` command; subcommand `tasks` routes to the
  tasks handler (sibling of `open`, `create-task`, `debug-go`).
- **Workspace finder** — starts at `--dir` or process cwd; walks parents until
  `<root>/.vscode/tasks.json` exists; root is the directory that contains
  `.vscode/`.
- **JSONC loader** — reads `tasks.json` allowing `//` line comments and trailing
  commas (VS Code / lifelog style).
- **Task model** — each entry has `label`, optional `type` (`process`|`shell`|
  empty), `command`/`args`, `options.cwd`, `isBackground`, `dependsOn` (string or
  list). Empty type/command with only dependsOn ⇒ **composite**.
- **Matcher** — exact label first; else unique case-insensitive substring;
  ambiguous / not found ⇒ error.
- **Plan expander** — for `run` (dry-run or execute): resolve dependsOn DAG
  (parallel siblings in the plan), expand `${workspaceFolder}`,
  `${workspaceFolderBasename}`, `${env:NAME}`; reject cycles, missing deps,
  unresolved `${…}`.
- **Backend selector** — `--backend=auto|local|iterm2` (default `auto`):
  - **local** — exec leaf command(s) in-process (sequential multi-step OK);
    background multi may warn on stderr with `warning:` prefix; exit code from
    last/failed child.
  - **iterm2** — map each plan **leaf** to a tab; **dry-run** prints tab plan
    offline (never calls RunTabSet); **live** (no `--dry-run`) calls
    `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2` **`RunTabSet`** with an
    ephemeral `TabSetSpec` (no `~/.config/iterm2/tab-set` write). **Fail closed:**
    RunTabSet / backend errors ⇒ non-zero exit; **no** silent local fallback and
    **no** stub warning `live iterm2 run not configured`.
  - **auto** — 1 non-background leaf → local; multi or background → iterm2 on
    darwin, else local (+ warning if multi/bg forced local).
- **RunTabSet injector (product seam)** — package-level func var and/or env mock
  so tests never open live iTerm:
  - `KOOL_VSCODE_TASKS_ITERM2_MOCK=1` — test-only: do **not** call real iTerm;
    record the intended `RunTabSet` call as JSON at
    `KOOL_VSCODE_TASKS_ITERM2_MOCK_OUT` and return success (unless mock-err).
  - `KOOL_VSCODE_TASKS_ITERM2_MOCK_ERR=1` — with mock: write call JSON (if out
    path set) then return error (fail-closed path for doctests).
  - Mock JSON (flexible; tests read common fields): `mode` string
    (`smart`|`new-window`|`no-new-window`), `spec.name`, `spec.windowName`,
    `spec.tabs[]` with `id`, `name`, `command`, `cwd`.
- **Tab mapping (locked defaults)**
  - Set **Name**: `vscode-tasks` + stable slug of root task label (e.g.
    `vscode-tasks-both-steps`).
  - **WindowName**: root task label.
  - Tab **Name**: leaf task label; Tab **ID**: stable slug of label (smart reuse).
  - Tab **Command** / **Cwd**: expanded plan leaf (same as dry-run).
  - Modes: default **smart**; `-n`/`--new-window` → **new-window**;
    `--no-new-window` → **no-new-window**.
- **Window flags** — `-n`/`--new-window` and `--no-new-window` are meaningful
  for iterm2; mutually exclusive (always error if both set). With
  `--backend=local`, these flags **error** (not ignored).
- **Caller** — supplies subcommand (`list`|`find`|`show`|`run`), optional query /
  label, flags (`--dir`, `--json`, `--dry-run`, `--backend`, `-n`,
  `--no-new-window`, `-h`/`--help`).

### Behaviors

**Commands**

```text
kool vscode tasks list [--dir <path>] [--json]
kool vscode tasks find <query> [--dir <path>]
kool vscode tasks show <label> [--dir <path>]
kool vscode tasks run  <label> [--dir <path>] [--dry-run]
    [--backend=auto|local|iterm2]
    [-n|--new-window] [--no-new-window]
kool vscode tasks -h|--help
```

- **help** — exit 0; usage mentions `tasks`, `list`, `find`, `show`, `run`,
  `--dry-run`, tasks.json / workspace (backend flags preferred).
- **list** — all tasks sorted by label; human table (LABEL, TYPE, BG, DEPS,
  COMMAND preview) + footer count and workspace path; `--json` ⇒ machine JSON,
  no ANSI. Missing tasks file after walk-up ⇒ error exit 1. Empty `tasks: []`
  ⇒ exit 0, zero count.
- **find** — case-insensitive substring on label; 0 matches ⇒ error; N matches
  ⇒ list exit 0.
- **show** — exact label preferred; else unique CI substring. Prints label, type,
  command/args, cwd, isBackground, dependsOn, workspace. Missing / ambiguous ⇒
  error.
- **run label matching** — same exact-then-unique-CI-substring rules; ambiguous
  lists matches; not found ⇒ error.
- **run --dry-run** — expand dependsOn; print plan (workspace, root task, steps
  with expanded vars). With `--backend=iterm2`, plan includes tab mapping
  offline. Cycle / missing dep / unresolved var ⇒ error. Never requires live
  iTerm; **never** invokes RunTabSet (mock out file must stay absent / empty).
- **run without --dry-run** — **actually execute** via selected backend.
- **run --backend=local** — spawn leaf command(s); stdout/stderr from children
  visible; multi-leaf sequential; composite expands deps then runs leaves.
- **run --backend=iterm2** (live) — call `RunTabSet` with mapped tabs; exit 0 on
  success; on error non-zero and **no** local multi-leaf fallback. Under mock
  env, record call JSON and skip real iTerm.
- **run --backend=auto** — heuristic above; single foreground leaf runs local
  offline-safely.
- **window flag conflict** — both `-n` and `--no-new-window` ⇒ error exit ≠ 0.
- **window flags + local** — error exit ≠ 0 (document: not ignored).
- **JSONC parse failure** — non-zero, message about invalid / parse.
- **Errors** — stderr with `Error:` (or kool sibling style), non-zero exit.
- **Trailing newline** after last content line on success.

**Variable expansion (plan and real run cwd/command)**

| Token | Value |
|-------|--------|
| `${workspaceFolder}` | workspace root (dir containing `.vscode`) |
| `${workspaceFolderBasename}` | basename of root |
| `${env:NAME}` | `os.Getenv("NAME")` |
| other `${…}` | error |

**Task kinds**

- `type: process` | `shell` with command → leaf
- no type/command, has `dependsOn` → composite
- `isBackground: true` → BG yes in list; influences auto → iterm2

**Source**

- Only `<workspace>/.vscode/tasks.json` (no user-level tasks, no write).
- Live iTerm not required in CI; live path covered via
  `KOOL_VSCODE_TASKS_ITERM2_MOCK*` env seam.

## Decision Tree

```
tasks/                              [nested DOCTEST root — P1 + P2 + live iterm2]
├── help/
│   └── show-usage/                 tasks -h / --help
├── list/
│   ├── missing-file/               no .vscode/tasks.json in walk-up
│   ├── empty-tasks/                valid file, tasks: []
│   ├── happy/                      multi-task + JSONC comments; sorted labels
│   ├── walk-up/                    --dir under nested cwd; finds parent tasks.json
│   └── json-flag/                  --json machine shape, no ANSI
├── find/
│   ├── unique/                     one CI substring match
│   ├── multi/                      several matches exit 0
│   └── zero/                       no match → error
├── show/
│   ├── composite/                  dependsOn-only task details
│   ├── leaf/                       shell/process leaf details
│   ├── missing/                    not found → error
│   └── ambiguous/                  multi CI match → error
├── run/
│   ├── dry-run/                    [P1 — plan only; stay GREEN]
│   │   ├── leaf/
│   │   ├── composite-deps/
│   │   ├── expand-workspaceFolder/
│   │   ├── unresolved-var/
│   │   ├── cycle/
│   │   └── missing-dep/
│   ├── match/                      [P1 — match rules via dry-run]
│   │   ├── exact/
│   │   ├── unique-substring/
│   │   ├── ambiguous/
│   │   └── not-found/
│   ├── exec-local/                 [P2 — --backend=local real exec]
│   │   ├── leaf-echo/              shell echo; stdout marker
│   │   ├── composite-sequential/   two leaf deps run; both markers
│   │   ├── expand-vars/            ${workspaceFolder}/env on real run
│   │   └── nonzero-exit/           failing command → exit ≠ 0
│   └── backend/                    [P2 — auto / iterm2 dry-run / live mock / flags]
│       ├── auto-single-local/      default auto; one fg leaf → local exec
│       ├── iterm2-dry-run-tabs/    --backend=iterm2 --dry-run multi tabs; no RunTabSet
│       ├── iterm2-live-invokes/    mock: RunTabSet once; map 2 tabs; no stub warning
│       ├── iterm2-mode-smart/      default mode smart in mock
│       ├── iterm2-mode-new-window/ -n → new-window in mock
│       ├── iterm2-mode-no-new-window/ --no-new-window → no-new-window
│       ├── iterm2-fail-no-local-fallback/ mock err → ≠0; no local multi fallback
│       ├── flag-conflict-windows/  -n and --no-new-window → error
│       └── local-with-new-window/  --backend=local -n → error
└── validation/
    └── invalid-jsonc/              broken tasks.json → error

# Retired (P1 seam): run/requires-dry-run — required error without --dry-run.
# Replaced by exec-local/* and backend/auto-single-local.
# Product stub to retire: "warning: live iterm2 run not configured; executing leaves locally"
```

## Test Index

| Leaf | Description | Expect |
|------|-------------|--------|
| `help/show-usage/` | Help exit 0; mentions tasks, list/find/show/run | P1 |
| `list/missing-file/` | No tasks.json after walk → exit ≠ 0 | P1 |
| `list/empty-tasks/` | Empty tasks array → exit 0, count 0 | P1 |
| `list/happy/` | JSONC multi-task list sorted; types/BG | P1 |
| `list/walk-up/` | Nested --dir finds parent workspace | P1 |
| `list/json-flag/` | `--json` machine output | P1 |
| `find/unique/` | One match exit 0 | P1 |
| `find/multi/` | Multi match exit 0 | P1 |
| `find/zero/` | Zero match error | P1 |
| `show/composite/` | Composite show fields | P1 |
| `show/leaf/` | Leaf show fields | P1 |
| `show/missing/` | Missing error | P1 |
| `show/ambiguous/` | Ambiguous error | P1 |
| `run/dry-run/leaf/` | Leaf dry-run plan exit 0 | P1 |
| `run/dry-run/composite-deps/` | Deps in plan | P1 |
| `run/dry-run/expand-workspaceFolder/` | Vars expanded | P1 |
| `run/dry-run/unresolved-var/` | Unresolved ${} error | P1 |
| `run/dry-run/cycle/` | Cycle error | P1 |
| `run/dry-run/missing-dep/` | Missing dep error | P1 |
| `run/match/exact/` | Exact label match | P1 |
| `run/match/unique-substring/` | Unique substring match | P1 |
| `run/match/ambiguous/` | Ambiguous match error | P1 |
| `run/match/not-found/` | Not found error | P1 |
| `run/exec-local/leaf-echo/` | `--backend=local` echo; stdout has marker | P2 |
| `run/exec-local/composite-sequential/` | Local sequential two leaves | P2 |
| `run/exec-local/expand-vars/` | Real run expands workspace/env | P2 |
| `run/exec-local/nonzero-exit/` | Failing leaf → exit ≠ 0 | P2 |
| `run/backend/auto-single-local/` | Default auto; one fg → local exec | P2 |
| `run/backend/iterm2-dry-run-tabs/` | iterm2 dry-run tab plan offline; mock not called | P2 |
| `run/backend/iterm2-live-invokes/` | mock RunTabSet once; 2 tabs; no stub warning | live RED |
| `run/backend/iterm2-mode-smart/` | live default → mode smart | live RED |
| `run/backend/iterm2-mode-new-window/` | `-n` → mode new-window | live RED |
| `run/backend/iterm2-mode-no-new-window/` | `--no-new-window` → mode no-new-window | live RED |
| `run/backend/iterm2-fail-no-local-fallback/` | mock err → ≠0; no local multi fallback | live RED |
| `run/backend/flag-conflict-windows/` | `-n` + `--no-new-window` error | P2 |
| `run/backend/local-with-new-window/` | local + `-n` error | P2 |
| `validation/invalid-jsonc/` | Parse error | P1 |

**P1 seam retired:** `run/requires-dry-run/` removed — real run is in scope for P2.

**Live iterm2 leaves (5):** RED until product wires `RunTabSet` + mock env seam
(`KOOL_VSCODE_TASKS_ITERM2_MOCK`, `_MOCK_OUT`, `_MOCK_ERR`). Mapping of tabs is
asserted on `iterm2-live-invokes` (no separate maps leaf).

## How to Run

```sh
# from kool module root
go build -o kool .
doctest vet ./tests/vscode/tasks
doctest test ./tests/vscode/tasks
```

P1/P2 leaves stay GREEN once product implements list/find/show/dry-run/local/auto.
Live iterm2 leaves stay RED until implementer replaces the local stub with
`RunTabSet` + mock env.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Request drives a single kool vscode tasks invocation (P1 + P2 + live mock).
type Request struct {
	// Phase: "cli" (subprocess kool binary). Handler phase may be added when
	// product exports RunForTest; default cli for Classic TDD.
	Phase string

	// Subcommand: list | find | show | run (empty when Help only).
	Subcommand string
	// Query is find <query> or show/run <label>.
	Query string

	// Flags
	Help   bool   // tasks -h / --help
	Dir    string // --dir <path> (workspace start)
	JSON   bool   // --json (list)
	DryRun bool   // --dry-run (run)

	// P2 run backends / window flags
	Backend     string // auto | local | iterm2 (empty → product default auto)
	NewWindow   bool   // -n / --new-window
	NoNewWindow bool   // --no-new-window

	ExtraArgs []string

	// WorkingDir is cmd.Dir (process cwd for walk-up when Dir empty).
	WorkingDir string

	// Env extras applied for dry-run / real-run ${env:NAME} tests and mock seam
	// (key=value). Prefer helpers enableITerm2Mock / enableITerm2MockErr.
	Env []string

	// ITerm2MockOut is set by enableITerm2Mock to the path product should write
	// mock RunTabSet call JSON (KOOL_VSCODE_TASKS_ITERM2_MOCK_OUT).
	ITerm2MockOut string
}

// Response is CLI capture after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// resolveKoolBinary finds a built kool for Phase=cli.
func resolveKoolBinary() (string, error) {
	if path, err := exec.LookPath("kool"); err == nil {
		return path, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		for _, rel := range []string{"kool", filepath.Join("bin", "kool")} {
			candidate := filepath.Join(dir, rel)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("kool binary not found; build with: go build -o kool .")
}

func buildTasksArgs(req *Request) []string {
	args := []string{"vscode", "tasks"}
	if req.Help {
		args = append(args, "--help")
		return args
	}
	if req.Subcommand != "" {
		args = append(args, req.Subcommand)
	}
	if req.Query != "" {
		args = append(args, req.Query)
	}
	if req.Dir != "" {
		args = append(args, "--dir", req.Dir)
	}
	if req.JSON {
		args = append(args, "--json")
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}
	if req.Backend != "" {
		args = append(args, "--backend", req.Backend)
	}
	if req.NewWindow {
		args = append(args, "-n")
	}
	if req.NoNewWindow {
		args = append(args, "--no-new-window")
	}
	args = append(args, req.ExtraArgs...)
	return args
}

func filterEnvWithout(keys ...string) []string {
	drop := map[string]bool{}
	for _, k := range keys {
		drop[k] = true
	}
	var out []string
	for _, e := range os.Environ() {
		key := e
		if i := strings.IndexByte(e, '='); i >= 0 {
			key = e[:i]
		}
		if drop[key] {
			continue
		}
		out = append(out, e)
	}
	return out
}

// Run invokes kool vscode tasks via CLI (Classic TDD — no in-process handler yet).
//
// Product surface to pin:
//
//	kool vscode tasks <subcommand> …
//	--dir, --json, --dry-run, --backend, -n/--new-window, --no-new-window, -h|--help
//
// Live iterm2 mock (product test-only hook):
//
//	KOOL_VSCODE_TASKS_ITERM2_MOCK=1
//	KOOL_VSCODE_TASKS_ITERM2_MOCK_OUT=<path>
//	KOOL_VSCODE_TASKS_ITERM2_MOCK_ERR=1
func Run(t *testing.T, req *Request) (*Response, error) {
	if req.Phase == "" {
		req.Phase = "cli"
	}
	args := buildTasksArgs(req)

	switch req.Phase {
	case "cli":
		koolBin, err := resolveKoolBinary()
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(koolBin, args...)
		if req.WorkingDir != "" {
			cmd.Dir = req.WorkingDir
		}
		// Drop mock keys from ambient env so leaf Env / helpers fully control seam.
		env := filterEnvWithout(
			"KOOL_VSCODE_TASKS_ITERM2_MOCK",
			"KOOL_VSCODE_TASKS_ITERM2_MOCK_OUT",
			"KOOL_VSCODE_TASKS_ITERM2_MOCK_ERR",
		)
		env = append(env, "PATH="+os.Getenv("PATH"))
		for _, e := range req.Env {
			env = append(env, e)
		}
		cmd.Env = env
		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		runErr := cmd.Run()
		exitCode := 0
		if runErr != nil {
			if ee, ok := runErr.(*exec.ExitError); ok {
				exitCode = ee.ExitCode()
			} else {
				return nil, fmt.Errorf("run kool: %w", runErr)
			}
		}
		return &Response{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: exitCode,
		}, nil
	default:
		return nil, fmt.Errorf("unknown phase %q (supports cli only)", req.Phase)
	}
}
```
