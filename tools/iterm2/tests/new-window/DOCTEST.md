# kool iterm2 — new-window flag (-n/--new-window) + --reuse-window alias

**DSN (Domain Specific Notion)**

```
Participants:
- User: invokes `kool iterm2 <dir> [flags]`
- CLI handler (tools/iterm2/iterm2.go): parses flags, builds Config, calls lib.OpenConfig
- Library (go-pkgs/shell/iterm2): builds AppleScript and runs osascript

Behaviors:
- User runs `kool iterm2 <dir>` → CLI parses flags → selects OpenMode → Library builds script → osascript runs
- ModeSmart (default): Library scans existing sessions, opens new tab in matching window or new window
- ModeReuseCurrent (-r/--reuse/--reuse-window): Library scans, focuses existing session or opens new window
- ModeForceNew (-n/--new-window): Library skips scan, always opens new window + cd
- -n/--new-window and -r/--reuse/--reuse-window are mutually exclusive: error reported to user
- --reuse-window is an alias for -r/--reuse, identical behavior
```

## Version

0.0.2

## Test tree

```
valid/                          # no conflict, valid flag combos
├── default/                    # no flags → ModeSmart (existing behavior)
├── reuse-r/                    # -r dir → ModeReuseCurrent (existing)
├── reuse-window-alias/         # --reuse-window dir → ModeReuseCurrent (new alias)
├── new-window-n/               # -n dir → ModeForceNew (new)
├── new-window-long/            # --new-window dir → ModeForceNew (new alias)
└── new-window-with-send/       # -n dir --send "echo hi" → ModeForceNew + followups
conflict/                       # -n and -r together
├── short-short/                # -n -r dir
├── long-long/                  # --new-window --reuse dir
├── mixed/                      # -n --reuse dir
└── reuse-window-conflict/      # -n --reuse-window dir
```

## How to run

```sh
doctest vet ./tools/iterm2/tests/new-window
doctest test ./tools/iterm2/tests/new-window
```

```go
import (
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "testing"

    iterm2cmd "github.com/xhd2015/kool/tools/iterm2"
)

// Request defines test inputs.
type Request struct {
    // Args is the CLI args for kool iterm2, excluding the dir.
    // e.g. ["-n", "--send", "echo hi"]
    // The Run function prepends the temp dir as the first arg.
    Args []string
}

// Response captures the outcome of running the CLI handler.
type Response struct {
    ExitCode int
    Stdout   string
    Stderr   string
    // ScriptText is the captured AppleScript (empty if handler error before library).
    ScriptText string
}

// Run executes the iterm2 CLI handler with mocked library dependencies.
func Run(t *testing.T, req *Request) (*Response, error) {
    tmpDir := t.TempDir()

    fullArgs := make([]string, 0, len(req.Args)+1)
    fullArgs = append(fullArgs, tmpDir)
    fullArgs = append(fullArgs, req.Args...)

    scriptOut := filepath.Join(tmpDir, "script.applescript")

    t.Setenv("KOOL_ITERM2_GOOS", "darwin")
    t.Setenv("KOOL_ITERM2_INSTALLED", "1")
    t.Setenv("KOOL_ITERM2_SCRIPT_OUT", scriptOut)

    var stdoutBuf, stderrBuf bytes.Buffer
    exitCode := iterm2cmd.RunForTest(fullArgs, &stdoutBuf, &stderrBuf, tmpDir)

    scriptText := ""
    if data, err := os.ReadFile(scriptOut); err == nil {
        scriptText = string(data)
    }

    return &Response{
        ExitCode:   exitCode,
        Stdout:     stdoutBuf.String(),
        Stderr:     stderrBuf.String(),
        ScriptText: scriptText,
    }, nil
}
```
