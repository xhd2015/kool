# kool for-every / for-every-\<duration\>

`kool for-every` runs a command in a loop with sleep between iterations, matching
bash `while true; do cmd; sleep N; done`. Duration may be a spaced positional or a
glued command suffix (`for-every-60s`). Shared duration parsing matches
`kool timeout` (via `pkgs/duration.Parse`). No storage; runtime options only.

## Version

0.0.2

## DSN (Domain Specific Notion)

### Participants

- **User** — invokes `kool for-every` or `kool for-every-<duration>` with optional
  stop/failure flags and a child command plus args.
- **kool CLI router** — `main.go` dispatches `for-every` and any
  `for-every-*` prefix to `tools/for-every`.
- **for-every handler** — parses flags (before the command), duration (positional
  or glued suffix), and the child command; validates; runs the loop.
- **duration parser** — shared `pkgs/duration.Parse`: prefer `time.ParseDuration`,
  else bare integer seconds; duration must be **> 0**.
- **Child process** — each iteration inherits stdin/stdout/stderr; exit status
  drives failure-policy counters.
- **Stop conditions** — `--max-runs`, effective max consecutive failures
  (`--max-failure` / `--allow-failure`), and signals (SIGINT/SIGTERM).

### Behaviors

- **Help** — `-h` / `--help` prints usage for both forms and flags; exit 0; no loop.
- **Validation (no loop)** — missing/invalid/non-positive duration, missing
  command, or invalid flag values → stderr error, non-zero exit, process ends.
- **Loop** — first run immediate; after each run, if still looping, sleep
  `Interval`; no overlapping runs.
- **Default on child failure** — log to stderr, continue; consecutive-failure
  counter increments; success resets it to 0.
- **`--allow-failure` alone** — treat as effective max consecutive failures = 1
  (exit on first child failure). Name is historical; help must document this.
- **`--max-failure N`** — stop after N consecutive failures (`N > 0`; omit/`0` =
  unlimited).
- **Both failure flags** — `--max-failure` wins (use its N).
- **`--max-runs N`** — stop after N iterations (success or fail each count);
  final exit 0 if last run succeeded, else last child exit code (prefer child
  code when available).
- **Stdout** — child output is passthrough; kool help/messages end lines with `\n`.

## Decision Tree

```
for-every/
├── help/                              [no loop; usage]
│   └── show-usage/                    -h/--help exit 0; both forms + flags
├── validation/                        [errors before loop]
│   ├── missing-duration/              spaced form, no duration positional
│   ├── invalid-duration/
│   │   ├── garbage/                   non-parseable string
│   │   └── non-positive/              0s / ≤0 rejected
│   ├── missing-command/
│   │   ├── spaced/                    for-every <dur> without command
│   │   └── glued/                     for-every-<dur> without command
│   └── max-runs-non-positive/         --max-runs 0 rejected
└── loop/                              [bounded loops; always stop flags]
    ├── spaced/                        for-every <duration> …
    │   ├── bare-int-duration/         duration "1" ≡ 1s; one run true
    │   ├── unit-duration-echo/        10ms + max-runs + echo
    │   ├── multi-arg-passthrough/     echo with multiple args
    │   └── max-runs-three/            exactly 3 successful iterations
    ├── glued/                         for-every-<duration> …
    │   └── unit-duration-echo/        same happy path as spaced unit form
    └── failure-policy/                child exit / flag interactions
        ├── default-continue/          fail + max-runs 3 → 3 attempts
        ├── allow-failure/             exit on first failure
        ├── max-failure-two/           stop after 2 consecutive fails
        ├── consec-reset/              success resets consecutive counter
        └── both-flags-max-wins/       allow-failure + max-failure 3 → 3
```

## Test Index

| Leaf | Description |
|------|-------------|
| `help/show-usage/` | `--help` exit 0; usage mentions both forms and failure/stop flags |
| `validation/missing-duration/` | Spaced `for-every` without duration → non-zero; no hang |
| `validation/invalid-duration/garbage/` | Invalid duration string → non-zero; message mentions duration |
| `validation/invalid-duration/non-positive/` | `0s` → non-zero; duration must be > 0 |
| `validation/missing-command/spaced/` | Valid duration, no command → non-zero |
| `validation/missing-command/glued/` | Glued form, no command → non-zero |
| `validation/max-runs-non-positive/` | `--max-runs 0` → non-zero validation |
| `loop/spaced/bare-int-duration/` | Bare `1` duration + `--max-runs 1 true` → exit 0 |
| `loop/spaced/unit-duration-echo/` | Spaced `10ms` + echo twice → exact stdout lines |
| `loop/spaced/multi-arg-passthrough/` | Multi-arg echo preserved on stdout |
| `loop/spaced/max-runs-three/` | `--max-runs 3` + printing cmd → exactly 3 lines, exit 0 |
| `loop/glued/unit-duration-echo/` | `for-every-10ms` + echo twice → same as spaced |
| `loop/failure-policy/default-continue/` | Always-fail + max-runs 3 → 3 attempts, non-zero |
| `loop/failure-policy/allow-failure/` | `--allow-failure` + fail → stop after 1, non-zero |
| `loop/failure-policy/max-failure-two/` | `--max-failure 2` + always fail → 2 runs then exit |
| `loop/failure-policy/consec-reset/` | F/S/F/S/F with max-failure 2 + max-runs 5 completes 5 |
| `loop/failure-policy/both-flags-max-wins/` | Both flags + always fail → 3 runs (max-failure wins) |

## How to Run

```sh
doctest vet ./tests/for-every
doctest test ./tests/for-every
```

`Run` builds `kool` once per process from the module root (see root `SETUP.md`)
so leaves exercise the workspace binary. Loop leaves always pass `--max-runs`
and/or `--max-failure` / `--allow-failure`; `Run` also applies a process
wall-clock timeout so a missing stop flag cannot hang the suite forever.

```go
import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

// Request drives a single kool for-every / for-every-<duration> invocation.
type Request struct {
	// Help runs `kool for-every --help` (ignores other fields except WorkingDir).
	Help bool

	// Glued selects for-every-<Duration>; false uses spaced for-every <Duration>.
	Glued bool

	// Duration is the interval string (positional or glued suffix). Empty when Help.
	Duration string

	// MaxRuns: nil omits the flag; non-nil passes --max-runs <value> (may be 0 for validation).
	MaxRuns *int
	// MaxFailure: nil omits; non-nil passes --max-failure <value>.
	MaxFailure *int
	// AllowFailure passes --allow-failure (exit on first child failure when MaxFailure unset).
	AllowFailure bool

	// Command and Args are the child process (after duration for spaced form).
	Command string
	Args    []string

	// WorkingDir is the kool process cwd (counter files, isolation). Set by root Setup.
	WorkingDir string

	// ProcessTimeout bounds the kool subprocess wall clock (default 15s).
	ProcessTimeout time.Duration
}

// Response is CLI capture after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Process-local kool binary (one-process suite; in-memory mutex, not session flock).
var (
	koolBinMu   sync.Mutex
	koolBinPath string
	koolBinErr  error
)

// ensureKoolBinary builds kool once per process into MkdirTemp (not t.TempDir).
func ensureKoolBinary(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	koolBinMu.Lock()
	defer koolBinMu.Unlock()
	if koolBinPath != "" || koolBinErr != nil {
		return koolBinPath, koolBinErr
	}
	dir, err := os.MkdirTemp("", "kool-for-every-doctest-bin-")
	if err != nil {
		koolBinErr = err
		return "", err
	}
	bin := filepath.Join(dir, "kool")
	moduleRoot := filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = moduleRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		koolBinErr = fmt.Errorf("go build kool: %w\n%s", err, out)
		return "", koolBinErr
	}
	koolBinPath = bin
	return bin, nil
}

func intPtr(n int) *int { return &n }

func buildArgs(req *Request) []string {
	if req.Help {
		return []string{"for-every", "--help"}
	}
	var args []string
	if req.Glued {
		args = append(args, "for-every-"+req.Duration)
	} else {
		args = append(args, "for-every")
	}
	// Flags before duration (spaced) / before command (glued), per CLI contract.
	if req.MaxRuns != nil {
		args = append(args, "--max-runs", strconv.Itoa(*req.MaxRuns))
	}
	if req.MaxFailure != nil {
		args = append(args, "--max-failure", strconv.Itoa(*req.MaxFailure))
	}
	if req.AllowFailure {
		args = append(args, "--allow-failure")
	}
	if !req.Glued {
		if req.Duration != "" {
			args = append(args, req.Duration)
		}
	}
	if req.Command != "" {
		args = append(args, req.Command)
		args = append(args, req.Args...)
	}
	return args
}

// Run executes kool for-every…, captures stdout/stderr/exit, and never hangs forever.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	koolBin, err := ensureKoolBinary(t, d)
	if err != nil {
		return nil, err
	}
	timeout := req.ProcessTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, koolBin, buildArgs(req)...)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if ctx.Err() == context.DeadlineExceeded {
		return resp, fmt.Errorf("kool for-every exceeded process timeout %v (missing stop flag?); stderr=%q", timeout, resp.Stderr)
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
			return resp, nil
		}
		return nil, fmt.Errorf("run kool: %w", runErr)
	}
	resp.ExitCode = 0
	return resp, nil
}
```
