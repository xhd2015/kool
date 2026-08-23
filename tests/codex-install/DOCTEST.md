# kool codex install — Ensure / dry-run / check-update

`kool codex install` wraps
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install` so the CLI can:

- **Ensure** (default) — install | update | noop
- **`--dry-run`** — plan only (no shell mutators)
- **`--check-update`** — report local vs latest status; exit 0 on success path
  even when outdated or missing

**Default layer: L2** in-process `tools/codex.RunForTest` with injectable
LookPath / RunShell / RunVersion / FetchLatest (via `SetDepsForTest`).
No real network, no real `codex` binary, no `curl | sh`.

**Classic TDD:** package `tools/codex` and `main.go` `case "codex"` do **not**
exist yet. Root `Run` imports them so the suite is **RED** (compile fail) until
the implementer lands the package + L2 inject surface.

**Out of scope (P2):** iterm2 changes; publishing; real npm / real install e2e.

## Version

0.0.1

## DSN (Domain Specific Notion)

### Participants

- **User** — invokes `kool codex install [flags]` or `kool codex --help` /
  `kool codex install --help`.
- **kool CLI router** — `main.go` dispatches reserved first arg `codex` to
  `tools/codex`.
- **`tools/codex`** — parses `install` subcommand and flags
  (`--dry-run`, `--check-update`, `-h|--help`); orchestrates Ensure / plan /
  status via library constants and injectable deps.
- **Library `shell/codex/install`** — `InstallCmd`, `UpdateCmd`, `Ensure`,
  version helpers. CLI must use library constants for shell commands (not
  re-literal the curl string in product code unless identical).
- **L2 inject** — package-level deps overridden per `Run` via
  `SetDepsForTest` (LookPath, RunShell, RunVersion, FetchLatest); suite holds a
  mutex around inject + call.

### Behaviors

- **Help (install)** — `kool codex install --help` documents `--dry-run` and
  `--check-update`; exit 0; stdout ends with `\n`. (Root `kool codex --help`
  also exists product-side; this suite seals install-level help.)
- **`--dry-run` (missing)** — LookPath miss → plan would **install** using
  `InstallCmd`; exit 0; **no** `RunShell`; **no** latest fetch required.
- **`--dry-run` (outdated)** — present + local < latest → plan would **update**
  using `UpdateCmd`; stdout shows local + latest versions; exit 0; no shell.
- **`--dry-run` (current)** — present + local == latest → plan **noop** / up to
  date; exit 0; no shell.
- **`--check-update` (outdated)** — status “update available” (or equivalent);
  exit **0**; no shell.
- **`--check-update` (current)** — status “up to date”; exit 0; no shell.
- **`--check-update` (missing)** — status “missing”; exit 0; no shell.
- **Ensure default (missing)** — `RunShell` exactly once with `InstallCmd`.
- **Ensure default (outdated)** — `RunShell` exactly once with `UpdateCmd`.
- **Ensure default (current)** — no shell mutator (noop).
- **Flag names** exactly `--dry-run` and `--check-update`.
- **Exit 0** on successful check-update paths even when outdated/missing.

### Expected implementer surface (for GREEN)

Package: `github.com/xhd2015/kool/tools/codex`

```go
// Deps are injectable for L2 doctests (nil fields → production defaults).
type Deps struct {
	LookPath    func(file string) (string, error)
	RunShell    func(ctx context.Context, cmd string) error
	RunVersion  func(ctx context.Context, bin string) (string, error)
	FetchLatest func(ctx context.Context) (string, error)
}

// SetDepsForTest installs package-level deps for one test invocation.
// Returns restore; Run always restores under a process mutex.
func SetDepsForTest(d Deps) (restore func())

// RunForTest runs the codex handler in-process (args after "kool codex").
// Example: RunForTest([]string{"install", "--dry-run"}, stdout, stderr)
func RunForTest(args []string, stdout, stderr io.Writer) int
```

Alternate acceptable names (if implementer prefers):
`SetCodexInstallDepsForTest` / `RunInstallCLI` — suite can be retargeted; prefer
the names above. Document any rename in a follow-up.

Library constants (import, do not re-spell):

```go
install.InstallCmd // curl -fsSL https://chatgpt.com/codex/install.sh | sh
install.UpdateCmd  // codex update
```

Recommended CLI routing:

```
kool codex install              → Ensure
kool codex install --dry-run    → plan only
kool codex install --check-update → status only, exit 0
kool codex install -h|--help    → install help
kool codex -h|--help            → codex root help
```

## Decision Tree

```
codex-install/
├── DOCTEST.md
├── SETUP.md
├── help/                                   [usage; exit 0]
│   ├── SETUP.md
│   └── show-usage/                         install --help lists --dry-run + --check-update
├── dry-run/                                [plan only; no shell]
│   ├── SETUP.md
│   ├── missing/                            would install InstallCmd; no shell; no latest fetch
│   ├── outdated/                           would update UpdateCmd + versions; no shell
│   └── current/                            noop / up to date; no shell
├── check-update/                           [status only; exit 0; no shell]
│   ├── SETUP.md
│   ├── outdated/                           update available; exit 0
│   ├── current/                            up to date; exit 0
│   └── missing/                            missing; exit 0
└── ensure/                                 [mutate via RunShell]
    ├── SETUP.md
    ├── missing-installs/                   ShellCalls == [InstallCmd]
    ├── outdated-updates/                   ShellCalls == [UpdateCmd]
    └── current-noop/                       ShellCalls empty
```

### Parameter significance (high → low)

1. **Mode** — help | dry-run | check-update | ensure
2. **Presence / version pair** — missing | outdated (0.1.0→0.2.0) | current (0.147.0)
3. **Shell spy** — InstallCmd | UpdateCmd | empty
4. **Exit code** — help/dry-run/check-update/ensure success → 0

## Test Index

| Leaf | Description | Classic |
|------|-------------|---------|
| `help/show-usage/` | `install --help` exit 0; includes `--dry-run` + `--check-update`; trailing `\n` | RED |
| `dry-run/missing/` | LookPath miss → would install `InstallCmd`; exit 0; no shell; FetchLatestCalls==0 | RED |
| `dry-run/outdated/` | present 0.1.0 + latest 0.2.0 → would update `UpdateCmd` + versions; no shell | RED |
| `dry-run/current/` | present current → noop/up to date; no shell | RED |
| `check-update/outdated/` | update available status; exit 0; no shell | RED |
| `check-update/current/` | up to date status; exit 0; no shell | RED |
| `check-update/missing/` | missing status; exit 0; no shell | RED |
| `ensure/missing-installs/` | shell `InstallCmd` once | RED |
| `ensure/outdated-updates/` | shell `UpdateCmd` once | RED |
| `ensure/current-noop/` | no shell | RED |

## How to Run

```sh
# from kool module root (external/kool-master-2026-08-10-1)
doctest vet ./tests/codex-install
doctest test ./tests/codex-install
doctest test -v ./tests/codex-install/help/show-usage
doctest test -v ./tests/codex-install/dry-run
doctest test -v ./tests/codex-install/check-update
doctest test -v ./tests/codex-install/ensure
```

Classic TDD: expect **RED** (compile failure until `tools/codex` exists, then
assert failures against incomplete implementations). Package inject is
process-global — this harness serializes inject + `RunForTest` with a mutex
(do not rely on leaf `t.Parallel()` isolation for those vars). No `t.Setenv` /
`t.Chdir` in leaves.

```go
import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xhd2015/doctest/session"
	codexinstall "github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install"
	codex "github.com/xhd2015/kool/tools/codex"
)

// codexInjectMu serializes package-level SetDepsForTest mutation.
var codexInjectMu sync.Mutex

// Request drives one in-process kool codex install invocation via RunForTest.
type Request struct {
	// Help: install --help (ignores mode flags except WorkingDir).
	Help bool

	// DryRun passes --dry-run.
	DryRun bool
	// CheckUpdate passes --check-update.
	CheckUpdate bool

	// Present: LookPath finds codex (false → missing bin).
	Present bool
	// LocalRaw is injected RunVersion stdout when Present (e.g. "codex-cli 0.1.0").
	LocalRaw string
	// Latest is injected FetchLatest success value (e.g. "0.2.0").
	Latest string
	// LatestFail makes FetchLatest return an error.
	LatestFail bool

	// BinPath is the resolved path returned by LookPath when Present.
	// Empty → WorkDir/bin/codex.
	BinPath string
	// WorkDir is an isolated temp root (root Setup).
	WorkDir string

	// WorkingDir reserved (RunForTest has no chdir contract for install).
	WorkingDir string
}

// Response is CLI capture + injection spies after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int

	ShellCalls       []string
	FetchLatestCalls int
	LookPathCalls    []string
	RunVersionCalls  []string
}

func buildInstallArgs(req *Request) []string {
	args := []string{"install"}
	if req.Help {
		args = append(args, "--help")
		return args
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}
	if req.CheckUpdate {
		args = append(args, "--check-update")
	}
	return args
}

// Run invokes tools/codex.RunForTest for install under inject mutex + SetDepsForTest.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}

	bin := req.BinPath
	if bin == "" {
		bin = filepath.Join(req.WorkDir, "bin", "codex")
	}

	resp := &Response{}
	args := buildInstallArgs(req)
	var stdout, stderr bytes.Buffer

	deps := codex.Deps{
		LookPath: func(file string) (string, error) {
			resp.LookPathCalls = append(resp.LookPathCalls, file)
			if !req.Present {
				return "", fmt.Errorf("lookpath: %s: not found", file)
			}
			return bin, nil
		},
		RunShell: func(ctx context.Context, cmd string) error {
			_ = ctx
			resp.ShellCalls = append(resp.ShellCalls, cmd)
			return nil
		},
		RunVersion: func(ctx context.Context, b string) (string, error) {
			_ = ctx
			resp.RunVersionCalls = append(resp.RunVersionCalls, b)
			out := req.LocalRaw
			if out == "" {
				out = "codex-cli 0.147.0"
			}
			return out, nil
		},
		FetchLatest: func(ctx context.Context) (string, error) {
			_ = ctx
			resp.FetchLatestCalls++
			if req.LatestFail {
				return "", fmt.Errorf("injected latest fetch failure")
			}
			v := req.Latest
			if v == "" {
				v = "0.147.0"
			}
			return v, nil
		},
	}

	codexInjectMu.Lock()
	defer codexInjectMu.Unlock()
	restore := codex.SetDepsForTest(deps)
	defer restore()

	code := codex.RunForTest(args, &stdout, &stderr)
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.ExitCode = code
	return resp, nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertNoShell(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.ShellCalls) != 0 {
		t.Fatalf("ShellCalls = %#v, want empty (no mutation)", resp.ShellCalls)
	}
}

func assertShellCalls(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ShellCalls = %#v, want %#v", got, want)
	}
	for i := range want {
		// Library Update/Install may path-qualify the binary when Bin is set
		// (e.g. "/tmp/.../bin/codex update"); accept that as matching "codex update".
		if shellCallMatches(got[i], want[i]) {
			continue
		}
		t.Fatalf("ShellCalls[%d] = %q, want %q", i, got[i], want[i])
	}
}

func shellCallMatches(got, want string) bool {
	if got == want {
		return true
	}
	gFields := strings.Fields(got)
	wFields := strings.Fields(want)
	if len(gFields) != len(wFields) || len(wFields) == 0 {
		return false
	}
	// Compare argv0 basenames; remaining args exact.
	if filepath.Base(gFields[0]) != filepath.Base(wFields[0]) {
		return false
	}
	for i := 1; i < len(wFields); i++ {
		if gFields[i] != wFields[i] {
			return false
		}
	}
	return true
}

// silence unused helpers in some leaves (constants still available for asserts)
var (
	_ = codexinstall.InstallCmd
	_ = codexinstall.UpdateCmd
	_ = strings.Contains
)
```
