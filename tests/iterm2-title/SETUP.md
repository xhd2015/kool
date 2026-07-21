# Scenario

**Feature**: kool iterm2 set-title / get-title CLI against fake osascript

```
# in-session success
kool iterm2 set-title|get-title [--window] … + ITERM_SESSION_ID
  -> shell/iterm2 title API -> fake osascript (script + optional stdout)

# not in iTerm2
kool iterm2 set-title|get-title … without ITERM_SESSION_ID
  -> stderr warning, exit 1 (no osascript)
```

## Preconditions

- `go build -o kool .` from module root (or `kool` on PATH).
- Implementer adds `set-title` / `get-title` dispatch under `tools/iterm2` and
  title helpers in `shell/iterm2` (session vs window, `ITERM_SESSION_ID`).
- Test hooks reused/extended: `KOOL_ITERM2_INSTALLED`, `KOOL_ITERM2_SCRIPT_OUT`,
  `KOOL_ITERM2_OSASCRIPT_EXIT`, plus `KOOL_ITERM2_OSASCRIPT_STDOUT` for mock gets.
- Fake `osascript` is placed on PATH ahead of the system binary per leaf.

## Steps

1. Root `Setup` creates temp working dir, `bin/`, and fake osascript.
2. Grouping/leaf `Setup` sets `Command`, session env, title, `--window`, mocks.
3. `Run` executes `kool iterm2 …` subprocess and captures stdout/stderr/script.

## Context

- Default session id: `w0t0p0:11111111-2222-3333-4444-555555555555`.
- Script capture path: `$WorkingDir/captured.applescript`.
- `InSession=false` strips `ITERM_SESSION_ID` from the child env.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	binDir := filepath.Join(req.WorkingDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	writeFakeOsascript(t, binDir)
	if req.InstalledEnv == "" {
		req.InstalledEnv = "1"
	}
	return nil
}

// markRootTree keeps hierarchical child packages importing this package live.
func markRootTree() {}
```
