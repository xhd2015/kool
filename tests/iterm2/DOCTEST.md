# kool iterm2 CLI

`kool iterm2 <dir> [-r] [--send <command>]...` opens a directory in iTerm2 on macOS using
the shared `shell/iterm2` library. Default smart-open scans session paths and opens a new
tab when a match exists. `-r` uses the same scan: on match it focuses the tab/session at
`targetDir` (no `cd`, no `--send`); on miss it opens a new window with `cd` and follow-ups.
Repeatable `--send` applies only on the miss path.

**Tab-set CLI** (`kool iterm2 tab-set …`) lives in the nested root
`./tests/iterm2/tab-set` (own `DOCTEST.md` / `Request` / `Run`) so open-dir leaves
stay independent.

## Version

0.0.2

## DSN (Domain Specific Notion)

### Participants

- **kool CLI** — `tools/iterm2` handler; parses `<dir>`, `-r`/`--reuse`, and `--send` flags.
- **`shell/iterm2`** — `OpenConfig` with AppleScript build and osascript execution.
- **Fake osascript** — test `osascript` on PATH writes script to
  `KOOL_ITERM2_SCRIPT_OUT` and honors `KOOL_ITERM2_OSASCRIPT_EXIT`.
- **Test env overrides** — `KOOL_ITERM2_INSTALLED` forces install check for subprocess tests.

### Behaviors

- **Validation** — missing `<dir>`, extra positionals, missing path, or non-directory
  → stderr + exit 1 before osascript.
- **Success (default)** — exit 0; captured script scans `path`, reuses via new tab or new window, `cd` + optional follow-ups.
- **Success (`-r`)** — exit 0; same scan; match branch focuses tab/session only; miss branch new window + `cd` + follow-ups.
- **Install / osascript errors** — stderr message + exit 1.
- **`--help`** — usage on stdout, exit 0.
- **Non-darwin** — exit 1 with macOS-only message (handler test with `SetGOOSForTest`).

## Decision Tree

```
iterm2/
├── validation/                 [CLI argv / path invalid]
│   ├── missing-arg/
│   ├── nonexistent-path/
│   ├── not-directory/
│   └── extra-args/
├── help/
│   └── show-usage/
├── cli/                        [successful subprocess + script capture]
│   ├── cd-only/
│   ├── match-branch-cd-scoped-to-window/
│   ├── smart-open-scans-user-variable/
│   ├── single-send/
│   ├── multiple-send/
│   └── reuse-flag/
│       ├── no-session-at-dir/
│       ├── session-at-dir/
│       ├── no-session-with-send/
│       ├── session-at-dir-with-send/
│       ├── miss-branch-registers-session/
│       ├── scan-matches-user-variable/
│       └── match-branch-selects-window/
└── error/
    ├── not-installed/
    ├── osascript-failure/
    └── unsupported-platform/
```

## Test Index

| Leaf | Description |
|------|-------------|
| `validation/missing-arg/` | No directory argument → usage on stderr |
| `validation/nonexistent-path/` | Missing path before osascript |
| `validation/not-directory/` | File path rejected |
| `validation/extra-args/` | Extra positional args rejected |
| `help/show-usage/` | `--help` prints usage, exit 0 |
| `cli/cd-only/` | Script has cd, no follow-up lines |
| `cli/match-branch-cd-scoped-to-window/` | Default mode: cd scoped to matchingWindow's new tab |
| `cli/smart-open-scans-user-variable/` | Default mode: scan matches `path` or `user.koolTargetDir` |
| `cli/single-send/` | `--send grok` in script |
| `cli/multiple-send/` | `--send grok --send codex` ordered |
| `cli/reuse-flag/no-session-at-dir/` | `-r` script scans paths; miss branch new window + cd |
| `cli/reuse-flag/session-at-dir/` | `-r` match branch focus tab/session only (no cd/tab create) |
| `cli/reuse-flag/no-session-with-send/` | `-r --send grok` → grok only in miss branch |
| `cli/reuse-flag/session-at-dir-with-send/` | `-r --send grok` → match branch suppresses grok |
| `cli/reuse-flag/miss-branch-registers-session/` | `-r` miss branch sets `user.koolTargetDir` for back-to-back reuse |
| `cli/reuse-flag/scan-matches-user-variable/` | `-r` scan matches `path` or `user.koolTargetDir` |
| `cli/reuse-flag/match-branch-selects-window/` | `-r` match branch selects matchingWindow to front |
| `error/not-installed/` | `KOOL_ITERM2_INSTALLED=0` → exit 1 |
| `error/osascript-failure/` | Fake osascript exit 1 → exit 1 |
| `error/unsupported-platform/` | `SetGOOSForTest(linux)` on handler |

## How to Run

```sh
doctest vet ./tests/iterm2
doctest test ./tests/iterm2
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

	iterm2cmd "github.com/xhd2015/kool/tools/iterm2"
)

const (
	koolIterm2InstalledEnv     = "KOOL_ITERM2_INSTALLED"
	koolIterm2ScriptOutEnv     = "KOOL_ITERM2_SCRIPT_OUT"
	koolIterm2OsascriptExitEnv = "KOOL_ITERM2_OSASCRIPT_EXIT"
)

type Request struct {
	Phase          string
	DirPath        string
	Reuse          bool
	Send           []string
	Help           bool
	ExtraPositional []string
	WorkingDir     string
	InstalledEnv   string
	OsascriptExit  int
	GoOS           string
}

type Response struct {
	Stdout       string
	Stderr       string
	ExitCode     int
	CapturedScript string
	HandlerErr   string
}

func resolveKoolBinary() (string, error) {
	moduleRoot := filepath.Join(DOCTEST_ROOT, "..", "..")
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
	body := `#!/bin/sh
if [ -n "$KOOL_ITERM2_SCRIPT_OUT" ]; then
  printf '%s' "$3" > "$KOOL_ITERM2_SCRIPT_OUT"
fi
exit "${KOOL_ITERM2_OSASCRIPT_EXIT:-0}"
`
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
}

func configureCLIEnv(t *testing.T, req *Request, cmd *exec.Cmd, scriptOut string) {
	t.Helper()
	env := os.Environ()
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
	if req.GoOS != "" {
		env = append(env, "KOOL_ITERM2_GOOS="+req.GoOS)
	} else {
		env = append(env, "KOOL_ITERM2_GOOS=darwin")
	}
	cmd.Env = env
}

func runCLI(t *testing.T, req *Request) (*Response, error) {
	koolBin, err := resolveKoolBinary()
	if err != nil {
		return nil, err
	}
	args := []string{"iterm2"}
	if req.Help {
		args = append(args, "--help")
	}
	if req.Reuse {
		args = append(args, "-r")
	}
	for _, s := range req.Send {
		args = append(args, "--send", s)
	}
	if req.DirPath != "" {
		args = append(args, req.DirPath)
	}
	args = append(args, req.ExtraPositional...)

	scriptOut := filepath.Join(req.WorkingDir, "captured.applescript")
	cmd := exec.Command(koolBin, args...)
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
	}
	return resp, nil
}

func runHandler(t *testing.T, req *Request) (*Response, error) {
	if req.GoOS != "" {
		iterm2cmd.SetGOOSForTest(req.GoOS)
		t.Cleanup(func() { iterm2cmd.SetGOOSForTest("") })
	}
	args := []string{}
	if req.Help {
		args = append(args, "--help")
	}
	if req.Reuse {
		args = append(args, "-r")
	}
	for _, s := range req.Send {
		args = append(args, "--send", s)
	}
	if req.DirPath != "" {
		args = append(args, req.DirPath)
	}
	args = append(args, req.ExtraPositional...)
	var stdout, stderr bytes.Buffer
	code := iterm2cmd.RunForTest(args, &stdout, &stderr, req.WorkingDir)
	resp := &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: code,
	}
	return resp, nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Phase {
	case "cli":
		return runCLI(t, req)
	case "handler":
		return runHandler(t, req)
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}
```