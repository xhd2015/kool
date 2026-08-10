# Scenario

**Feature**: kool codex install (Ensure / --dry-run / --check-update)

```
# help
user -> kool codex install --help
  -> usage includes --dry-run + --check-update; exit 0

# dry-run (injected LookPath / versions; no shell)
user -> kool codex install --dry-run
  -> plan install | update | noop; exit 0; ShellCalls empty

# check-update (status only)
user -> kool codex install --check-update
  -> status missing | update available | up to date; exit 0; no shell

# ensure (default)
user -> kool codex install
  -> RunShell InstallCmd | UpdateCmd | none
```

## Preconditions

- Module root is `DOCTEST_ROOT/../..` (this tree lives at `tests/codex-install/`).
- Package under test: `github.com/xhd2015/kool/tools/codex`
  (`RunForTest`, `SetDepsForTest` / `Deps`).
- Library (via go.mod replace):  
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install`  
  (`InstallCmd`, `UpdateCmd`).
- **Classic TDD:** `tools/codex` does not exist yet → suite is **RED** (import
  compile failure) until implementer lands package + main dispatch.
- L2 only: injectable LookPath / RunShell / RunVersion / FetchLatest.
  No real network, no real `codex` binary, no process env/cwd mutation.
- Package inject is process-global: root `Run` holds `codexInjectMu` around
  inject + call (leaves must not assume parallel isolation of those vars).
- No `t.Setenv` / `t.Chdir` in leaves.

## Steps

1. Root `Setup` allocates `WorkDir` under `t.TempDir()`.
2. Grouping `Setup` sets mode flags (Help / DryRun / CheckUpdate).
3. Leaf `Setup` sets presence + version fixtures.
4. Root `Run` builds `install …` args, injects deps, calls `RunForTest`,
   returns stdout/stderr/exit + shell spies.

## Context

- Outdated fixture: local raw `codex-cli 0.1.0`, latest `0.2.0`.
- Current fixture: local raw `codex-cli 0.147.0`, latest `0.147.0`.
- Missing fixture: `Present=false` (LookPath miss); FetchLatest must not be
  required for dry-run/check-update/ensure install path.
- Install recipe constant:  
  `curl -fsSL https://chatgpt.com/codex/install.sh | sh` (`install.InstallCmd`)
- Update command constant: `codex update` (`install.UpdateCmd`)
- Help stdout must end with trailing `\n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.WorkDir = t.TempDir()
	req.Help = false
	req.DryRun = false
	req.CheckUpdate = false
	req.Present = false
	req.LocalRaw = ""
	req.Latest = ""
	req.LatestFail = false
	req.BinPath = ""
	return nil
}
```
