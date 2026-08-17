# Scenario

**Feature**: kool iterm2 CLI opens directories in iTerm2 with optional follow-ups

```
# CLI subprocess path
kool iterm2 <dir> [--send cmd]... -> shell/iterm2.OpenConfig -> fake osascript (tests)

# handler path (platform errors)
RunForTest(args) -> parse flags -> OpenConfig
```

## Preconditions

- `go build -o kool .` from module root (or `kool` on PATH).
- Implementer adds `tools/iterm2` with `RunForTest`, `SetGOOSForTest`, and env hooks:
  `KOOL_ITERM2_INSTALLED`, `KOOL_ITERM2_SCRIPT_OUT`, `KOOL_ITERM2_OSASCRIPT_EXIT`.
- Fake `osascript` is placed on PATH ahead of system binary in each CLI leaf.

## Steps

1. Root `Setup` creates temp working dir, `bin/`, and fake osascript.
2. Leaves set `req.Phase`, paths, flags, and env overrides.
3. `Run` executes CLI subprocess or in-process handler per phase.

## Context

- Relative `<dir>` resolves against `req.WorkingDir` (leaf `chdir` not required; CLI uses cwd).
- Script capture file: `$WorkingDir/captured.applescript`.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	binDir := filepath.Join(req.WorkingDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	writeFakeOsascript(t, binDir)
	if req.Phase == "" {
		req.Phase = "cli"
	}
	return nil
}

func initValidDir(t *testing.T, base, name string) string {
	t.Helper()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// writeTextCmd matches shell/iterm2 SafeInputIgnore write text "cmd" lines (Ctrl-U prefix).
func writeTextCmd(cmd string) string {
	return `write text ((ASCII character 21) & "` + cmd + `")`
}

// writeTextCd matches the Ctrl-U-prefixed cd line emitted for session commands.
func writeTextCd() string {
	return `write text ((ASCII character 21) & "cd " & quoted form of targetDir)`
}
```