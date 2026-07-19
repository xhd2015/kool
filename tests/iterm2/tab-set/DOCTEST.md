# kool iterm2 tab-set CLI

`kool iterm2 tab-set` lists, shows, runs, status-checks, and stops named tab-set
configs under `~/.config/iterm2/tab-set` (overridable via
`KOOL_ITERM2_TAB_SET_DIR` for tests). Open-dir / title CLIs are covered by
sibling trees (`./tests/iterm2`, `./tests/iterm2-title`).

## Version

0.0.2

**Classic TDD (this cycle):** config-mode leaves (list/show/run dry-run/flag
conflict/validation/help) stay GREEN. New ad-hoc `--tab` and `--save` leaves
under `run/adhoc/`, `run/save/`, and related help/flag leaves are intentionally
RED until the implementer extends `tab-set run` (no new subcommand).

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
- **Orchestrator** — `shell/iterm2` `RunTabSet` / `StatusTabSet` / `StopTabSet`
  (production); CLI **dry-run** and **--save** must not require live iTerm.
- **Caller** — supplies subcommand, set name, flags (`--dry-run`, `-n`,
  `--no-new-window`, `--tab`, `--save`, `--force`, `--window-name`).

### Behaviors

**Commands**

```text
kool iterm2 tab-set list
kool iterm2 tab-set show <name>
kool iterm2 tab-set run <name> [flags]
kool iterm2 tab-set status <name>
kool iterm2 tab-set stop <name>
kool iterm2 tab-set -h|--help
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
  `--save` → error.
- **Overwrite:** TTY → prompt y/N after nice diff; non-TTY without `--force` →
  error; decline → non-zero, no write.
- **Diff buckets:** unchanged / modified / added / deleted (+ window_name).
- **`-n` / `--no-new-window` with `--save`:** error (unused with save-only).
- **`--window-name`:** optional for ad-hoc; stored on save.

**`--tab` syntax**

```text
spaces [ spaces props spaces ] spaces command
```

- Arbitrary spaces before `[` and after `]`.
- props: `key=value` comma-separated; keys: `id`, `name`, `cwd`.
- Leading `[` that fails props parse → **error**.
- No props → entire trimmed string is command.
- Default id: `tab-1` … `tab-N` (1-based `--tab` order) if no `id=`.
- Default name: same as id; cwd empty if omitted.
- Duplicate ids → error.

**Other commands (config mode)**

- **help** — exit 0; stdout mentions `tab-set`, subcommands, config path; after
  this cycle also `--tab` / `--save`.
- **list** — exit 0; empty dir → empty list or “0 sets”; with fixtures → names.
- **show** — prints window name + tab ids/commands; missing → error.
- **run --dry-run** (config) — exit 0; plan; no iTerm.
- **run -n + --no-new-window** — error, exit 1.
- **validation** — version ≠ 1, duplicate ids, empty tabs → error.

**Config JSON (version 1)**

```json
{
  "version": 1,
  "window_name": "local-bots",
  "tabs": [
    {"id": "a", "name": "a", "command": "echo a"},
    {"id": "b", "name": "b", "command": "echo b", "cwd": "/tmp"}
  ]
}
```

- `id` optional → default from `name`
- reject version != 1, empty tabs, duplicate ids, missing command

**Env**

- `KOOL_ITERM2_TAB_SET_DIR` — absolute path to config directory (required in tests).

## Decision Tree

```
tab-set/                            [nested DOCTEST root]
├── help/
│   ├── show-usage/                 tab-set -h / --help (existing)
│   └── adhoc-flags/                help mentions --tab / --save / --force  [RED]
├── list/
│   ├── empty/
│   └── one-set/
├── show/
│   ├── bots/
│   └── missing/
├── run/
│   ├── dry-run/                    config mode --dry-run
│   ├── flag-conflict/              -n + --no-new-window
│   ├── force-without-save/         --force without --save → error          [RED]
│   ├── adhoc/                      ≥1 --tab, no --save
│   │   ├── dry-run-tabs/           two tabs, --dry-run; no config file
│   │   ├── default-tab-ids/        bare commands → tab-1, tab-2
│   │   ├── props-whitespace/       "  [ id = a ]  echo hi"
│   │   ├── invalid-props/          bad [...] → error
│   │   └── dup-id/                 duplicate id → error
│   └── save/                       --save (never iTerm)
│       ├── create/                 write new JSON; no iTerm
│       ├── dry-run-create/         no write; create plan
│       ├── overwrite-force/        --force overwrite + diff buckets
│       ├── overwrite-non-tty-no-force/  exists, no force → error
│       ├── without-tab/            --save alone → error
│       └── rejects-new-window-flag/ --save -n → error
└── validation/
    ├── bad-version/
    ├── dup-ids/
    └── empty-tabs/
```

## Test Index

| Leaf | Description | Expect |
|------|-------------|--------|
| `help/show-usage/` | Help exit 0; mentions tab-set, list/run, config | GREEN |
| `help/adhoc-flags/` | Help mentions `--tab`, `--save`, `--force` | RED |
| `list/empty/` | Empty config dir → exit 0, empty / 0 sets | GREEN |
| `list/one-set/` | Fixture `bots` listed with tab count | GREEN |
| `show/bots/` | Shows tab ids and commands | GREEN |
| `show/missing/` | Unknown set → error, exit ≠ 0 | GREEN |
| `run/dry-run/` | config `--dry-run` plan; exit 0 | GREEN |
| `run/flag-conflict/` | `-n` + `--no-new-window` → error | GREEN |
| `run/force-without-save/` | `--force` without `--save` → error | RED |
| `run/adhoc/dry-run-tabs/` | `--tab`×2 `--dry-run`; no config file | RED |
| `run/adhoc/default-tab-ids/` | bare cmds → tab-1, tab-2 in plan | RED |
| `run/adhoc/props-whitespace/` | spaced props block parses | RED |
| `run/adhoc/invalid-props/` | bad `[...]` → error | RED |
| `run/adhoc/dup-id/` | duplicate ad-hoc id → error | RED |
| `run/save/create/` | `--save` writes v1 JSON; no iTerm | RED |
| `run/save/dry-run-create/` | `--save --dry-run` no write | RED |
| `run/save/overwrite-force/` | `--save --force` overwrite + diff | RED |
| `run/save/overwrite-non-tty-no-force/` | non-TTY no force → error | RED |
| `run/save/without-tab/` | `--save` alone → error | RED |
| `run/save/rejects-new-window-flag/` | `--save -n` → error | RED |
| `validation/*` | schema validation | GREEN |

## How to Run

```sh
# from kool module root
go build -o kool .
doctest vet ./tests/iterm2/tab-set
doctest test ./tests/iterm2/tab-set
```

Expect mixed: existing config-mode leaves GREEN, new ad-hoc/save leaves RED.

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
	"testing"

	iterm2cmd "github.com/xhd2015/kool/tools/iterm2"
)

const (
	// TabSetDirEnv overrides default ~/.config/iterm2/tab-set for tests.
	TabSetDirEnv = "KOOL_ITERM2_TAB_SET_DIR"
)

// Request drives a single kool iterm2 tab-set invocation.
type Request struct {
	// Phase: "handler" (in-process RunForTest, preferred) or "cli" (subprocess).
	Phase string

	// Subcommand: list | show | run | status | stop (empty when Help only).
	Subcommand string
	// SetName is the config basename (e.g. "bots") for show/run/status/stop.
	SetName string

	// Flags for run (and help).
	Help        bool // tab-set -h / --help
	DryRun      bool // --dry-run
	NewWindow   bool // -n / --new-window
	NoNewWindow bool // --no-new-window

	// Ad-hoc / save flags (Classic TDD this cycle — product not yet wired).
	// Tabs: each entry is one --tab value (command and optional [props]).
	Tabs       []string
	Save       bool   // --save
	Force      bool   // --force (with --save only)
	WindowName string // --window-name <name>

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

func applyTabSetEnv(t *testing.T, req *Request) {
	t.Helper()
	if req.ConfigDir != "" {
		prev, had := os.LookupEnv(TabSetDirEnv)
		if err := os.Setenv(TabSetDirEnv, req.ConfigDir); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
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
	_ = os.Setenv("KOOL_ITERM2_GOOS", goos)
	t.Cleanup(func() {
		if hadGOOS {
			_ = os.Setenv("KOOL_ITERM2_GOOS", prevGOOS)
		} else {
			_ = os.Unsetenv("KOOL_ITERM2_GOOS")
		}
	})
}

// Run prefers in-process tools/iterm2.RunForTest (Phase=handler).
// Phase=cli uses the kool binary with the same argv and env.
//
// Product surface to pin (Classic TDD):
//
//	kool iterm2 tab-set <subcommand> …
//	env KOOL_ITERM2_TAB_SET_DIR
//	RunForTest([]string{"tab-set", …}, stdout, stderr, workingDir) int
func Run(t *testing.T, req *Request) (*Response, error) {
	if req.Phase == "" {
		req.Phase = "handler"
	}
	applyTabSetEnv(t, req)
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
