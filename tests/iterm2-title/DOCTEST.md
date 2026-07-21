# kool iterm2 set-title / get-title CLI

`kool iterm2 set-title [--window] <title>` and `kool iterm2 get-title [--window]`
read and write the current iTerm2 session or window title when the process is
rooted in an iTerm2 session (`ITERM_SESSION_ID` set). Open-dir behavior of
`kool iterm2 <dir> …` is out of scope for this tree (covered by `./tests/iterm2`).

## Version

0.0.2

## DSN (Domain Specific Notion)

### Participants

- **kool CLI** — `tools/iterm2` handler; dispatches reserved first args
  `set-title` / `get-title` / help before open-dir routing.
- **`shell/iterm2` title API** — detects in-session via `ITERM_SESSION_ID`,
  builds AppleScript to get/set session `name` (default) or window name
  (`--window`), runs osascript.
- **Fake osascript** — test binary on PATH; writes script body to
  `KOOL_ITERM2_SCRIPT_OUT`, optional stdout from `KOOL_ITERM2_OSASCRIPT_STDOUT`,
  exit from `KOOL_ITERM2_OSASCRIPT_EXIT`.
- **Session env** — `ITERM_SESSION_ID` (e.g. `w0t0p0:UUID`); empty/missing means
  not inside iTerm2.

### Behaviors

- **Not in iTerm2** — no osascript; stderr warning
  (`nothing to set` / `nothing to get` … `inside iTerm2`); exit 1.
- **set-title validation** — missing or empty `<title>` → stderr error, exit 1.
- **set-title success** — get current title, set new; stdout
  `title changed: <old> -> <new>\n` (empty old printed as-is); exit 0.
  Default target = session/tab name; `--window` = containing window title.
- **get-title success** — print current title + trailing `\n`; exit 0.
- **get-title extra args** — exit 1.
- **Help** — `kool iterm2 --help` lists open-dir usage and both title subcommands.
- **Script capture** — success paths leave a non-empty captured AppleScript that
  references the session UUID and the chosen target (session vs window).

## Decision Tree

```
iterm2-title/
├── set-title/                      [subcommand = set-title]
│   ├── not-in-iterm2/              ITERM_SESSION_ID empty
│   ├── missing-title/              no title positional
│   ├── empty-title/                title == ""
│   ├── osascript-failure/          in-session; fake osascript exit 1
│   ├── session/                    default target (tab/session name)
│   │   ├── happy/                  old title mock → success message
│   │   ├── empty-old-title/        mock old empty → "title changed:  -> …"
│   │   └── escaping/               title with " and \
│   └── window/                     --window target
│       ├── happy/
│       └── flag-after-title/       title then --window
├── get-title/                      [subcommand = get-title]
│   ├── not-in-iterm2/
│   ├── extra-args/                 get-title foo
│   ├── session/
│   │   └── happy/
│   └── window/
│       └── happy/
└── help/
    └── lists-title-cmds/           --help mentions set-title + get-title
```

## Test Index

| Leaf | Description |
|------|-------------|
| `set-title/not-in-iterm2/` | Unset session → exit 1; warning; no script |
| `set-title/missing-title/` | `set-title` alone → exit 1 validation |
| `set-title/empty-title/` | `set-title ""` → exit 1 |
| `set-title/osascript-failure/` | Valid set; osascript exit 1 → CLI exit 1 |
| `set-title/session/happy/` | Session name set; stdout `title changed: old -> new\n` |
| `set-title/session/empty-old-title/` | Empty old title in success line |
| `set-title/session/escaping/` | Quotes/backslashes in title; success + script escapes |
| `set-title/window/happy/` | `--window` sets window title |
| `set-title/window/flag-after-title/` | `set-title <title> --window` accepted |
| `get-title/not-in-iterm2/` | Unset session → exit 1; nothing-to-get warning |
| `get-title/extra-args/` | Extra positional → exit 1 |
| `get-title/session/happy/` | Prints session title + `\n` |
| `get-title/window/happy/` | `--window` prints window title + `\n` |
| `help/lists-title-cmds/` | Help lists set-title and get-title |

## How to Run

```sh
doctest vet ./tests/iterm2-title
doctest test ./tests/iterm2-title
```

Build kool first: `go build -o kool .`

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	koolIterm2InstalledEnv      = "KOOL_ITERM2_INSTALLED"
	koolIterm2ScriptOutEnv      = "KOOL_ITERM2_SCRIPT_OUT"
	koolIterm2OsascriptExitEnv  = "KOOL_ITERM2_OSASCRIPT_EXIT"
	koolIterm2OsascriptStdoutEnv = "KOOL_ITERM2_OSASCRIPT_STDOUT"
	defaultTestSessionID        = "w0t0p0:11111111-2222-3333-4444-555555555555"
)

// Request drives a single kool iterm2 title subcommand (or help).
type Request struct {
	// Command is "set-title", "get-title", or empty when Help is true.
	Command string
	// Title is the set-title positional; TitleSet distinguishes missing vs "".
	Title    string
	TitleSet bool
	// Window requests --window (window title vs session name).
	Window bool
	// WindowAfterTitle places --window after the title positional.
	WindowAfterTitle bool
	// ExtraArgs are additional positionals (e.g. get-title extras).
	ExtraArgs []string
	// Help runs kool iterm2 --help.
	Help bool
	// InSession controls ITERM_SESSION_ID injection (true = set, false = strip).
	InSession bool
	// SessionID overrides defaultTestSessionID when InSession.
	SessionID string
	// WorkingDir is the per-leaf temp dir (bin/, captured script).
	WorkingDir string
	// InstalledEnv sets KOOL_ITERM2_INSTALLED (default "1").
	InstalledEnv string
	// OsascriptExit sets KOOL_ITERM2_OSASCRIPT_EXIT when non-zero.
	OsascriptExit int
	// OsascriptStdout is printed by fake osascript (mock get-title / old title).
	OsascriptStdout string
	// GoOS sets KOOL_ITERM2_GOOS (default darwin).
	GoOS string
}

// Response is CLI capture after Run.
type Response struct {
	Stdout         string
	Stderr         string
	ExitCode       int
	CapturedScript string
	ScriptWritten  bool
}

func resolveKoolBinary(d *session.Doctest) (string, error) {
	moduleRoot := filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	candidates := []string{
		filepath.Join(moduleRoot, "kool"),
		filepath.Join(moduleRoot, "bin", "kool"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	if path, err := exec.LookPath("kool"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("kool binary not found; build with: go build -o kool .")
}

func writeFakeOsascript(t *testing.T, binDir string) {
	t.Helper()
	script := filepath.Join(binDir, "osascript")
	// Capture -e SCRIPT ($3), optional stdout mock, optional non-zero exit.
	body := `#!/bin/sh
if [ -n "$KOOL_ITERM2_SCRIPT_OUT" ]; then
  printf '%s' "$3" > "$KOOL_ITERM2_SCRIPT_OUT"
fi
if [ -n "$KOOL_ITERM2_OSASCRIPT_STDOUT" ]; then
  printf '%s' "$KOOL_ITERM2_OSASCRIPT_STDOUT"
fi
exit "${KOOL_ITERM2_OSASCRIPT_EXIT:-0}"
`
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
}

func sessionID(req *Request) string {
	if req.SessionID != "" {
		return req.SessionID
	}
	return defaultTestSessionID
}

func sessionUUID(req *Request) string {
	sid := sessionID(req)
	if i := strings.Index(sid, ":"); i >= 0 && i+1 < len(sid) {
		return sid[i+1:]
	}
	return sid
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

func configureCLIEnv(t *testing.T, req *Request, cmd *exec.Cmd, scriptOut string) {
	t.Helper()
	env := filterEnvWithout(
		"ITERM_SESSION_ID",
		"PATH",
		koolIterm2InstalledEnv,
		koolIterm2ScriptOutEnv,
		koolIterm2OsascriptExitEnv,
		koolIterm2OsascriptStdoutEnv,
		"KOOL_ITERM2_GOOS",
	)
	binDir := filepath.Join(req.WorkingDir, "bin")
	env = append(env, "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	installed := req.InstalledEnv
	if installed == "" {
		installed = "1"
	}
	env = append(env, koolIterm2InstalledEnv+"="+installed)
	if scriptOut != "" {
		env = append(env, koolIterm2ScriptOutEnv+"="+scriptOut)
	}
	if req.OsascriptExit != 0 {
		env = append(env, fmt.Sprintf("%s=%d", koolIterm2OsascriptExitEnv, req.OsascriptExit))
	}
	if req.OsascriptStdout != "" {
		env = append(env, koolIterm2OsascriptStdoutEnv+"="+req.OsascriptStdout)
	}
	if req.GoOS != "" {
		env = append(env, "KOOL_ITERM2_GOOS="+req.GoOS)
	} else {
		env = append(env, "KOOL_ITERM2_GOOS=darwin")
	}
	if req.InSession {
		env = append(env, "ITERM_SESSION_ID="+sessionID(req))
	}
	cmd.Env = env
}

func buildArgs(req *Request) []string {
	args := []string{"iterm2"}
	if req.Help {
		args = append(args, "--help")
		return args
	}
	if req.Command != "" {
		args = append(args, req.Command)
	}
	if req.Window && !req.WindowAfterTitle {
		args = append(args, "--window")
	}
	if req.Command == "set-title" && req.TitleSet {
		args = append(args, req.Title)
	}
	if req.Window && req.WindowAfterTitle {
		args = append(args, "--window")
	}
	args = append(args, req.ExtraArgs...)
	return args
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	koolBin, err := resolveKoolBinary(d)
	if err != nil {
		return nil, err
	}
	scriptOut := filepath.Join(req.WorkingDir, "captured.applescript")
	_ = os.Remove(scriptOut)

	cmd := exec.Command(koolBin, buildArgs(req)...)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}
	configureCLIEnv(t, req, cmd, scriptOut)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("run kool: %w", runErr)
		}
	}
	resp := &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
	if b, err := os.ReadFile(scriptOut); err == nil {
		resp.CapturedScript = string(b)
		resp.ScriptWritten = true
	}
	return resp, nil
}
```
