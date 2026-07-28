# kool iterm2 tab-set CLI

`kool iterm2 tab-set` lists, shows, runs, status-checks, stops, and **updates**
named tab-set configs under `~/.config/iterm2/tab-set` (overridable via
`KOOL_ITERM2_TAB_SET_DIR` for tests). Open-dir / title CLIs are covered by
sibling trees (`./tests/iterm2`, `./tests/iterm2-title`).

## Version

0.0.3

**Classic TDD (this cycle — `update`):** fine-grained config mutation via
`kool iterm2 tab-set update <name> …` (patch fields, `--rm`, `--create`,
`--window-name`). New leaves under `update/**` are intentionally **RED** until
implementer lands. Prior leaves (list/show/run/save/no_submit/…) must stay
**GREEN**. Config-only — never iTerm / never `RunTabSet`. Schema stays
**version 1**.

## DSN (Domain Specific Notion)

### Participants

- **kool CLI** — `tools/iterm2` handler; reserved first arg `tab-set` before
  open-dir routing.
- **Config store** — JSON files `KOOL_ITERM2_TAB_SET_DIR/<name>.json` (default
  `~/.config/iterm2/tab-set/<name>.json`).
- **Config loader** — parse version-1 schema; validate tabs (ids, commands).
- **Tab parser** — ad-hoc `--tab` strings → tab specs (optional props + command).
- **Save planner** — compare ad-hoc tabs to existing JSON; print create/diff plan;
  optional confirm / `--force`; write file (never iTerm).
- **Update mutator** — load named set → apply field patch / create tab / remove
  tab / set `window_name` → validate → atomic write (never iTerm).
- **Orchestrator** — `shell/iterm2` `RunTabSet` / `StatusTabSet` / `StopTabSet`
  (production); CLI **dry-run**, **--save**, and **update** must not require live
  iTerm.
- **Caller** — supplies subcommand, set name, flags (`--dry-run`, `-n`,
  `--no-new-window`, `--tab`, `--save`, `--force`, `--window-name`,
  `--tab-id`, `--rm`, `--command`, `--name`, `--cwd`, `--clear-cwd`,
  `--no-submit`, `--submit`, `--create`).

### Behaviors

**Commands**

```text
kool iterm2 tab-set list
kool iterm2 tab-set show <name>
kool iterm2 tab-set run <name> [flags]
kool iterm2 tab-set update <name> [flags]
kool iterm2 tab-set status <name>
kool iterm2 tab-set stop <name>
kool iterm2 tab-set -h|--help
kool iterm2 tab-set update -h|--help
```

**run modes (locked)**

| Invocation | iTerm run? | Disk write? |
|------------|------------|-------------|
| `run name` (0 `--tab`) | yes | no |
| `run name --dry-run` (0 `--tab`) | no (run plan) | no |
| `run name --tab …` | yes | no |
| `run name --tab … --dry-run` | no (run plan) | no |
| `run name --tab … --save` | **no** | **yes** (after confirm) |
| `run name --tab … --save --dry-run` | **no** | **no** (save plan only) |
| `run name --save` without `--tab` | error | no |

- **≥1 `--tab`** → **ad-hoc mode**: do not read config JSON; file not required.
- **0 `--tab`** → **config mode**: load `<name>.json`.
- **`--save` never runs iTerm.** To run after save: second `run <name>` config mode.
- **`--force`:** with `--save`, skip y/N; still print diff on overwrite. Without
  `--save` → error (except `update --rm`, which also accepts `--force`).
- **Overwrite:** TTY → prompt y/N after nice diff; non-TTY without `--force` →
  error; decline → non-zero, no write.
- **Diff buckets:** unchanged / modified / added / deleted (+ window_name).
- **`-n` / `--no-new-window` with `--save`:** error (unused with save-only).
- **`--window-name`:** optional for ad-hoc; stored on save; also on `update`.

**update modes (locked — Classic TDD this cycle)**

| Invocation | iTerm? | Disk write? |
|------------|--------|-------------|
| `update name --tab-id id --no-submit` (etc.) | **no** | **yes** (immediate) |
| `update name --window-name W` (no `--tab-id`) | **no** | **yes** |
| `update name --tab-id id --rm --force` | **no** | **yes** |
| `update name --tab-id id --create --command …` | **no** | **yes** |
| `update name … --dry-run` | **no** | **no** (plan only) |

- **`--tab-id` required** for tab field patch, `--rm`, or `--create`.
- **`--window-name` alone** (no `--tab-id`) is a valid set-level patch.
- **`--rm` exclusive** with patch flags (`--command`, `--name`, `--cwd`,
  `--clear-cwd`, `--no-submit`, `--submit`, `--create`).
- **`--no-submit` / `--submit`** mutually exclusive.
- **`--cwd` / `--clear-cwd`** mutually exclusive.
- **At least one action** required; else “nothing to update”.
- **`--create`:** requires `--command`; tab id must not already exist.
  Without `--create`, missing id → error (hint `--create`).
- **`--rm` last remaining tab** → error (schema forbids empty `tabs`).
- **`--rm` confirm:** TTY → y/N unless `--force`; non-TTY without `--force` →
  error (same pattern as save overwrite). Handler `RunForTest` is non-TTY →
  success-path rm tests use `--force`.
- **Field patch:** no confirm; short change summary on stdout; write immediately.
- **Atomic write:** load → mutate → validate → write. Schema version 1.
- **Never runs iTerm** / never calls `RunTabSet`.
- Error prefix style: `tab-set update: …` on stderr (match other tab-set errors).

**`--tab` syntax**

```text
spaces [ spaces props spaces ] spaces command
```

- Arbitrary spaces before `[` and after `]`.
- props: `key=value` comma-separated; keys: `id`, `name`, `cwd`, `no_submit`.
- `no_submit` bool: `true`/`false`, `1`/`0`, `yes`/`no` (case-insensitive);
  other values → **error**.
- Leading `[` that fails props parse → **error**.
- No props → entire trimmed string is command.
- Default id: `tab-1` … `tab-N` (1-based `--tab` order) if no `id=`.
- Default name: same as id; cwd empty if omitted; no_submit false if omitted.
- Duplicate ids → error.

**Other commands (config mode)**

- **help** — exit 0; stdout mentions `tab-set`, subcommands, config path;
  `--tab` / `--save` / `no_submit`; update help mentions `--tab-id` / `--rm` /
  `--no-submit`.
- **list** — exit 0; empty dir → empty list or “0 sets”; with fixtures → names.
- **show** — prints window name + tab ids/commands; tabs with `no_submit` surface
  that flag; missing → error.
- **run --dry-run** (config) — exit 0; plan; no iTerm; mark no_submit tabs
  (e.g. `(no_submit)`).
- **run -n + --no-new-window** — error, exit 1.
- **update** — config-only mutation (see update modes above).
- **validation** — version ≠ 1, duplicate ids, empty tabs → error.

**Config JSON (version 1)**

```json
{
  "version": 1,
  "window_name": "local-bots",
  "tabs": [
    {"id": "a", "name": "a", "command": "echo a"},
    {"id": "b", "name": "b", "command": "grok --resume", "cwd": "/tmp", "no_submit": true}
  ]
}
```

- `id` optional → default from `name`
- `no_submit` optional bool; omit or false → submit (Enter); true → stage only
- reject version != 1, empty tabs, duplicate ids, missing command

**Env**

- `KOOL_ITERM2_TAB_SET_DIR` — absolute path to config directory (required in tests).

## Decision Tree

```
tab-set/                            [nested DOCTEST root]
├── help/
│   ├── show-usage/                 tab-set -h / --help (existing)
│   └── adhoc-flags/                help --tab/--save/--force + no_submit
├── list/
│   ├── empty/
│   └── one-set/
├── show/
│   ├── bots/                       omit no_submit still works (compat)
│   ├── missing/
│   └── no-submit/                  fixture no_submit=true surfaces in show
├── run/
│   ├── dry-run/                    config mode --dry-run (default submit)
│   ├── dry-run-no-submit/          plan marks no_submit tabs
│   ├── flag-conflict/              -n + --no-new-window
│   ├── force-without-save/         --force without --save → error
│   ├── adhoc/                      ≥1 --tab, no --save
│   │   ├── dry-run-tabs/           two tabs, --dry-run; no config file
│   │   ├── default-tab-ids/        bare commands → tab-1, tab-2
│   │   ├── props-whitespace/       "  [ id = a ]  echo hi"
│   │   ├── invalid-props/          bad [...] → error
│   │   ├── dup-id/                 duplicate id → error
│   │   ├── no-submit-true/         [id=x,no_submit=true] dry-run plan
│   │   └── invalid-no-submit/      no_submit=maybe → error
│   └── save/                       --save (never iTerm)
│       ├── create/                 write new JSON; no iTerm
│       ├── dry-run-create/         no write; create plan
│       ├── overwrite-force/        --force overwrite + diff buckets
│       ├── overwrite-non-tty-no-force/  exists, no force → error
│       ├── without-tab/            --save alone → error
│       ├── rejects-new-window-flag/ --save -n → error
│       ├── persist-no-submit/      JSON persists no_submit=true
│       └── overwrite-no-submit/    no_submit change → file/diff modified
├── update/                         update <name> (config-only; never iTerm) [RED]
│   ├── help/                       update -h; key flags
│   ├── patch/                      existing tab field ops
│   │   ├── no-submit/              --no-submit → JSON true
│   │   ├── submit/                 --submit clears no_submit
│   │   ├── command/                --command only
│   │   ├── name-cwd/               --name and/or --cwd
│   │   └── clear-cwd/              --clear-cwd → empty cwd
│   ├── window-name-only/           --window-name without --tab-id
│   ├── create/
│   │   ├── with-command/           --create --command appends tab
│   │   ├── needs-command/          --create without --command → error
│   │   └── already-exists/         --create on existing id → error
│   ├── rm/
│   │   ├── force/                  --rm --force removes tab
│   │   ├── last-tab/               last tab → error
│   │   ├── non-tty-no-force/       non-TTY rm without --force → error
│   │   └── with-patch-flags/       --rm + patch flags → exclusive error
│   ├── dry-run-no-write/           --dry-run plan; file unchanged
│   └── errors/
│       ├── missing-set/
│       ├── missing-tab/            no --create; id missing → error
│       ├── no-submit-and-submit/   exclusive
│       ├── cwd-and-clear-cwd/      exclusive
│       ├── nothing-to-update/      no action flags
│       └── tab-fields-without-tab-id/
└── validation/
    ├── bad-version/
    ├── dup-ids/
    └── empty-tabs/
```

## Test Index

| Leaf | Description | Expect |
|------|-------------|--------|
| `help/show-usage/` | Help exit 0; mentions tab-set, list/run, config | GREEN |
| `help/adhoc-flags/` | Help mentions `--tab`/`--save`/`--force` + `no_submit` | GREEN* |
| `list/empty/` | Empty config dir → exit 0, empty / 0 sets | GREEN |
| `list/one-set/` | Fixture `bots` listed with tab count | GREEN |
| `show/bots/` | Shows tab ids and commands (omit no_submit compat) | GREEN |
| `show/missing/` | Unknown set → error, exit ≠ 0 | GREEN |
| `show/no-submit/` | show surfaces `no_submit=true` for that tab | GREEN* |
| `run/dry-run/` | config `--dry-run` plan; exit 0 | GREEN |
| `run/dry-run-no-submit/` | plan marks no_submit tabs | GREEN* |
| `run/flag-conflict/` | `-n` + `--no-new-window` → error | GREEN |
| `run/force-without-save/` | `--force` without `--save` → error | GREEN |
| `run/adhoc/dry-run-tabs/` | `--tab`×2 `--dry-run`; no config file | GREEN |
| `run/adhoc/default-tab-ids/` | bare cmds → tab-1, tab-2 in plan | GREEN |
| `run/adhoc/props-whitespace/` | spaced props block parses | GREEN |
| `run/adhoc/invalid-props/` | bad `[...]` → error | GREEN |
| `run/adhoc/dup-id/` | duplicate ad-hoc id → error | GREEN |
| `run/adhoc/no-submit-true/` | ad-hoc `no_submit=true` dry-run plan | GREEN* |
| `run/adhoc/invalid-no-submit/` | `no_submit=maybe` → error | GREEN* |
| `run/save/create/` | `--save` writes v1 JSON; no iTerm | GREEN |
| `run/save/dry-run-create/` | `--save --dry-run` no write | GREEN |
| `run/save/overwrite-force/` | `--save --force` overwrite + diff | GREEN |
| `run/save/overwrite-non-tty-no-force/` | non-TTY no force → error | GREEN |
| `run/save/without-tab/` | `--save` alone → error | GREEN |
| `run/save/rejects-new-window-flag/` | `--save -n` → error | GREEN |
| `run/save/persist-no-submit/` | save persists `"no_submit": true` | GREEN* |
| `run/save/overwrite-no-submit/` | overwrite sets no_submit; file proves | GREEN* |
| `update/help/` | `update -h`; mentions `--tab-id`/`--rm`/`--no-submit` | RED |
| `update/patch/no-submit/` | existing tab → `"no_submit": true`; others unchanged | RED |
| `update/patch/submit/` | `--submit` clears no_submit | RED |
| `update/patch/command/` | only command changes | RED |
| `update/patch/name-cwd/` | name and/or cwd patch | RED |
| `update/patch/clear-cwd/` | cwd empty after | RED |
| `update/window-name-only/` | no `--tab-id`; only `window_name` changes | RED |
| `update/create/with-command/` | `--create --command` appends tab | RED |
| `update/create/needs-command/` | `--create` without `--command` → error | RED |
| `update/create/already-exists/` | `--create` on existing id → error | RED |
| `update/rm/force/` | `--rm --force`; other tabs remain | RED |
| `update/rm/last-tab/` | last remaining tab → error | RED |
| `update/rm/non-tty-no-force/` | non-TTY rm without `--force` → error | RED |
| `update/rm/with-patch-flags/` | `--rm` + patch exclusive error | RED |
| `update/dry-run-no-write/` | dry-run plan; file unchanged | RED |
| `update/errors/missing-set/` | unknown set name → error | RED |
| `update/errors/missing-tab/` | unknown tab id without `--create` | RED |
| `update/errors/no-submit-and-submit/` | exclusive flags → error | RED |
| `update/errors/cwd-and-clear-cwd/` | exclusive flags → error | RED |
| `update/errors/nothing-to-update/` | no action → error | RED |
| `update/errors/tab-fields-without-tab-id/` | tab patch without `--tab-id` | RED |
| `validation/*` | schema validation | GREEN |

\* no_submit leaves from prior cycle; expected GREEN if already implemented.

## How to Run

```sh
# from kool module root
go build -o kool .
doctest vet ./tests/iterm2/tab-set
doctest test ./tests/iterm2/tab-set
```

Expect mixed: existing leaves GREEN; new `update/**` leaves RED until implementer.

Parent open-dir suite (unaffected):

```sh
doctest test ./tests/iterm2
```

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xhd2015/doctest/session"
	iterm2cmd "github.com/xhd2015/kool/tools/iterm2"
)

const (
	// TabSetDirEnv overrides default ~/.config/iterm2/tab-set for tests.
	TabSetDirEnv = "KOOL_ITERM2_TAB_SET_DIR"
)

// tabSetTestEnvMu serializes process env mutations in Run. doctest generates
// t.Parallel() per leaf; concurrent os.Setenv of KOOL_ITERM2_TAB_SET_DIR would
// otherwise race (wrong ConfigDir / home default).
var tabSetTestEnvMu sync.Mutex

// Request drives a single kool iterm2 tab-set invocation.
type Request struct {
	// Phase: "handler" (in-process RunForTest, preferred) or "cli" (subprocess).
	Phase string

	// Subcommand: list | show | run | update | status | stop (empty when Help only).
	Subcommand string
	// SetName is the config basename (e.g. "bots") for show/run/update/status/stop.
	SetName string

	// Flags for run/update (and help).
	Help        bool // tab-set [-h|--help] or tab-set <sub> --help when Subcommand set
	DryRun      bool // --dry-run
	NewWindow   bool // -n / --new-window
	NoNewWindow bool // --no-new-window

	// Ad-hoc / save flags.
	// Tabs: each entry is one --tab value (command and optional [props]).
	Tabs       []string
	Save       bool   // --save
	Force      bool   // --force (with --save overwrite, or update --rm)
	WindowName string // --window-name <name> (run --save and update)

	// Update flags (subcommand "update"). Convention: string fields empty =
	// flag not provided; bool flags false = not provided.
	//   update <SetName> [--tab-id …] [--rm|--command|--name|--cwd|--clear-cwd|
	//   --no-submit|--submit|--create] [--window-name …] [--dry-run] [--force]
	TabID          string // --tab-id <id>
	Rm             bool   // --rm
	Command        string // --command <str> (non-empty = provided)
	TabName        string // --name <str> (non-empty = provided; tab display name)
	Cwd            string // --cwd <path> (non-empty = provided)
	ClearCwd       bool   // --clear-cwd
	UpdateNoSubmit bool   // --no-submit → no_submit: true
	UpdateSubmit   bool   // --submit → no_submit: false / omit
	Create         bool   // --create (requires Command; tab must not exist)

	ExtraArgs []string

	// ConfigDir is the absolute path set as KOOL_ITERM2_TAB_SET_DIR.
	ConfigDir string
	// WorkingDir is per-leaf temp workspace.
	WorkingDir string

	// GoOS for KOOL_ITERM2_GOOS (default darwin).
	GoOS string
}

// Response is CLI capture after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// resolveKoolBinary finds a built kool for Phase=cli. Avoids free DOCTEST_ROOT
// (session inject is scoped to Run/Setup/Assert; package-level helpers cannot
// close over it under current doctest assembly).
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

func buildTabSetArgs(req *Request) []string {
	args := []string{"tab-set"}
	if req.Help {
		// tab-set --help  OR  tab-set <subcommand> --help (e.g. update -h)
		if req.Subcommand != "" {
			args = append(args, req.Subcommand, "--help")
			return args
		}
		args = append(args, "--help")
		return args
	}
	if req.Subcommand != "" {
		args = append(args, req.Subcommand)
	}
	if req.SetName != "" {
		args = append(args, req.SetName)
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}
	if req.NewWindow {
		args = append(args, "-n")
	}
	if req.NoNewWindow {
		args = append(args, "--no-new-window")
	}
	for _, tab := range req.Tabs {
		args = append(args, "--tab", tab)
	}
	if req.Save {
		args = append(args, "--save")
	}
	if req.Force {
		args = append(args, "--force")
	}
	if req.WindowName != "" {
		args = append(args, "--window-name", req.WindowName)
	}
	// update flags
	if req.TabID != "" {
		args = append(args, "--tab-id", req.TabID)
	}
	if req.Rm {
		args = append(args, "--rm")
	}
	if req.Command != "" {
		args = append(args, "--command", req.Command)
	}
	if req.TabName != "" {
		args = append(args, "--name", req.TabName)
	}
	if req.Cwd != "" {
		args = append(args, "--cwd", req.Cwd)
	}
	if req.ClearCwd {
		args = append(args, "--clear-cwd")
	}
	if req.UpdateNoSubmit {
		args = append(args, "--no-submit")
	}
	if req.UpdateSubmit {
		args = append(args, "--submit")
	}
	if req.Create {
		args = append(args, "--create")
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

// applyTabSetEnv sets process env for the current Run call. Caller must hold
// tabSetTestEnvMu and restore env before releasing the lock (do not use
// t.Cleanup — it races with other parallel tests after unlock).
func applyTabSetEnv(t *testing.T, req *Request) (restore func()) {
	t.Helper()
	var restores []func()
	if req.ConfigDir != "" {
		prev, had := os.LookupEnv(TabSetDirEnv)
		if err := os.Setenv(TabSetDirEnv, req.ConfigDir); err != nil {
			t.Fatalf("set %s: %v", TabSetDirEnv, err)
		}
		restores = append(restores, func() {
			if had {
				_ = os.Setenv(TabSetDirEnv, prev)
			} else {
				_ = os.Unsetenv(TabSetDirEnv)
			}
		})
	}
	goos := req.GoOS
	if goos == "" {
		goos = "darwin"
	}
	prevGOOS, hadGOOS := os.LookupEnv("KOOL_ITERM2_GOOS")
	if err := os.Setenv("KOOL_ITERM2_GOOS", goos); err != nil {
		t.Fatalf("set KOOL_ITERM2_GOOS: %v", err)
	}
	restores = append(restores, func() {
		if hadGOOS {
			_ = os.Setenv("KOOL_ITERM2_GOOS", prevGOOS)
		} else {
			_ = os.Unsetenv("KOOL_ITERM2_GOOS")
		}
	})
	return func() {
		for i := len(restores) - 1; i >= 0; i-- {
			restores[i]()
		}
	}
}

// Run prefers in-process tools/iterm2.RunForTest (Phase=handler).
// Phase=cli uses the kool binary with the same argv and env.
//
// Product surface to pin (Classic TDD):
//
//	kool iterm2 tab-set <subcommand> …
//	env KOOL_ITERM2_TAB_SET_DIR
//	RunForTest([]string{"tab-set", …}, stdout, stderr, workingDir) int
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	if req.Phase == "" {
		req.Phase = "handler"
	}

	// Serialize env for process-wide KOOL_ITERM2_* vars under t.Parallel().
	tabSetTestEnvMu.Lock()
	defer tabSetTestEnvMu.Unlock()
	restore := applyTabSetEnv(t, req)
	defer restore()

	args := buildTabSetArgs(req)

	switch req.Phase {
	case "handler":
		var stdout, stderr bytes.Buffer
		code := iterm2cmd.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		return &Response{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: code,
		}, nil
	case "cli":
		koolBin, err := resolveKoolBinary()
		if err != nil {
			return nil, err
		}
		full := append([]string{"iterm2"}, args...)
		cmd := exec.Command(koolBin, full...)
		if req.WorkingDir != "" {
			cmd.Dir = req.WorkingDir
		}
		env := filterEnvWithout(TabSetDirEnv, "KOOL_ITERM2_GOOS", "PATH")
		env = append(env, "PATH="+os.Getenv("PATH"))
		if req.ConfigDir != "" {
			env = append(env, TabSetDirEnv+"="+req.ConfigDir)
		}
		goos := req.GoOS
		if goos == "" {
			goos = "darwin"
		}
		env = append(env, "KOOL_ITERM2_GOOS="+goos)
		cmd.Env = env
		var stdout, stderr bytes.Buffer
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
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}
```
