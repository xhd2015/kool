# kool sandbox events (Topology B + notify-event + hot reload)

**PHASE-1 / P0.** While a sealed sandbox runner has a live guest, it binds a
per-session Unix socket under `$KOOL_SANDBOX_ROOT/events/<session-id>.sock` and
accepts JSON events. `kool sandbox notify-event` is a **stateless publisher**
that dials every `*.sock` under the events dir (no broker). Sessions filter
`devbox.updated` and re-apply **only that load’s file layer** when `path` is in
the session’s runtime-load set (sealed `RuntimeLoadDevbox` + ad-hoc
`--load-devbox` at start).

**Classic TDD:** product surfaces (`notify-event` subcommand, runner event
listener, runtime-load file hot-reload) are **not implemented yet**. Leaves are
**RED** until the implementer lands them. This tree uses the session-built
`kool` binary (CLI) so the suite **compiles**; failures are assert / non-zero
exit / missing sock (not compile-only).

**Out of scope (this phase):** `remote-devbox refresh`, launchd, remote deploy,
seatalk restart, topology A broker, `sandbox.bin refresh` CLI, guest env
hot-reload, primary pack reload.

## Version

0.0.2

## DSN (Domain Specific Notion)

### Participants

- **User / publisher** — invokes
  `kool sandbox notify-event --type TYPE --path ABS [--root DIR] [--dry-run] [-h]`.
- **kool CLI router** — `main.go` → `tools/sandbox.Handle`; new subcommand
  `notify-event`.
- **Sealed runner (outer process)** — after materialize, creates
  `$KOOL_SANDBOX_ROOT/events/` (0700) and binds
  `$KOOL_SANDBOX_ROOT/events/<session-id>.sock` (0600) where session-id is the
  basename of the materialize root; concurrent accept loop while `cmd.Run`
  waits; on session end: close listener, unlink sock, then existing
  `RemoveAll` session.
- **Event bus (Topology B)** — no central broker; publisher dials each sock;
  stale sock → skip (optional unlink); no subscribers → warning on stderr,
  exit 0.
- **Runtime-load set** — absolute paths from sealed `RuntimeLoadDevbox` (pack
  order) then ad-hoc `--load-devbox` at session start (first-seen dedupe). Only
  these paths are eligible for hot file-layer reload.
- **Reload target** — on `devbox.updated` with absolute `path` in the set:
  re-unseal that path from disk and re-apply **only that load’s Files** into
  the existing session root (safe writes / later-wins overlay for that load).
  Unknown path → discard. Env is **not** hot-reloaded this phase.

### Event JSON

```json
{"v":1,"type":"devbox.updated","path":"/abs/path","ts":"<RFC3339>"}
```

### Behaviors

- **notify-event help** — `--help` / `-h` documents `--type`, `--path`,
  optional `--root`, `--dry-run`; exit 0; stdout ends with `\n`.
- **notify-event no socks** — events dir empty (or missing with empty list) →
  warning on stderr; exit **0**; no panic.
- **notify-event dry-run** — lists would-be sock targets (existing `*.sock`
  under events) without requiring live listeners; exit 0; does not require
  successful dial.
- **notify-event delivers** — with a live listener on a sock under
  `$ROOT/events/`, dials and sends JSON with `v`, `type`, `path`, `ts`;
  success summary counts on stdout; exit 0.
- **live session binds sock** — sealed primary with `--runtime-load-devbox`
  (or ad-hoc load) + long-lived guest; after start,
  `$KOOL_SANDBOX_ROOT/events/<session-id>.sock` exists; events dir mode 0700;
  sock mode 0600 (when inspectable).
- **hot reload files** — guest session has old load file content on disk under
  session root; rewrite/rebuild the load seal at the same abs path; publish
  `devbox.updated` for that path; **session root** file content becomes new
  **without restarting the guest**.
- **filter discard** — notify path not in session loads → session files
  unchanged.
- **session end cleans sock** — after guest exits, sock is unlinked.

### Expected implementer surface (for GREEN)

Product may expose library helpers (names flexible if docs updated):

```text
EventsDir(sandboxRoot string) string
  → filepath.Join(sandboxRoot, "events")

NotifyEvent / PublishDevboxUpdated(root, eventType, path string, dryRun bool) error
  → list EventsDir(root)/*.sock; dial each; send JSON; summary on stdout

ReloadRuntimeLoadFiles(sessionRoot, loadAbsPath string) error
  → re-unseal loadAbsPath; re-apply Files only into sessionRoot

Runner: always-on (or opt-in) event listener after materialize succeeds
```

**Default layer L2.** Parallel-safe: no `t.Setenv` / `t.Chdir` / global stdout
hijack; inject `KOOL_SANDBOX_ROOT` via **child cmd.Env** only. Unique temp dirs
per leaf via `d`/root Setup.

### CLI surfaces

```text
kool sandbox notify-event --type TYPE --path ABS [--root DIR] [--dry-run] [-h]
  --type   event type (this feature: devbox.updated)
  --path   absolute path (load seal path for devbox.updated)
  --root   sandbox root parent (default: env KOOL_SANDBOX_ROOT or product default)
  --dry-run list targets without requiring successful dials

./sandbox.bin [--load-devbox ABS]... [--] <long-lived guest>
  env KOOL_SANDBOX_ROOT=<parent>
  → materialize under <parent>/<session-id>/;
    bind <parent>/events/<session-id>.sock;
    accept loop + guest; cleanup sock then session
```

## Decision Tree

```
sandbox-events/
├── DOCTEST.md
├── SETUP.md
├── notify-event/                           [publisher CLI; L2]
│   ├── SETUP.md
│   ├── help/
│   │   ├── SETUP.md
│   │   └── documents-flags/                --type, --path; exit 0
│   ├── no-socks/
│   │   ├── SETUP.md
│   │   └── empty-events-dir/               warning + exit 0
│   ├── dry-run/
│   │   ├── SETUP.md
│   │   └── lists-targets/                  dry-run lists socks; exit 0
│   └── delivers/
│       ├── SETUP.md
│       └── fake-listener-receives-json/    mock unix sock gets event JSON
└── live-session/                           [live sealed runner + events; L2]
    ├── SETUP.md                            # NOT named session/ — collides with doctest session pkg
    ├── binds-sock/
    │   ├── SETUP.md
    │   └── runtime-load-live/              events/<id>.sock exists while guest runs
    ├── hot-reload/
    │   ├── SETUP.md
    │   └── files-update-without-restart/   rewrite load seal + notify → new file content
    ├── filter/
    │   ├── SETUP.md
    │   └── discard-unknown-path/           unknown path → session file unchanged
    └── cleanup/
        ├── SETUP.md
        └── sock-unlinked-on-exit/          after guest exit, sock gone
```

## Test Index

| Leaf | Description | Classic |
|------|-------------|---------|
| `notify-event/help/documents-flags/` | `notify-event --help` exit 0; documents `--type` + `--path`; trailing `\n` | RED |
| `notify-event/no-socks/empty-events-dir/` | empty events dir → warning stderr; exit 0 | RED |
| `notify-event/dry-run/lists-targets/` | `--dry-run` lists existing sock basename(s); exit 0; no dial required | RED |
| `notify-event/delivers/fake-listener-receives-json/` | mock listener receives `devbox.updated` JSON with path | RED |
| `live-session/binds-sock/runtime-load-live/` | live guest + runtime-load → `events/<id>.sock` exists | RED |
| `live-session/hot-reload/files-update-without-restart/` | rewrite load + notify → session file new content | RED |
| `live-session/filter/discard-unknown-path/` | notify other abs path → session file stays old | RED |
| `live-session/cleanup/sock-unlinked-on-exit/` | after guest exit, sock unlinked | RED |

## How to Run

```sh
# from kool module root
# If a leftover tests/sandbox-events/session/ tree exists (name collides with
# github.com/xhd2015/doctest/session), remove it first:
#   rm -rf tests/sandbox-events/session
doctest vet ./tests/sandbox-events
doctest test ./tests/sandbox-events
doctest test -v ./tests/sandbox-events/notify-event
doctest test -v ./tests/sandbox-events/live-session
```

Classic TDD: expect **RED** until `notify-event` + runner event bus +
runtime-load file reload land. Pre-existing helpers (session `kool` build,
secondary pack build) may pass; feature leaves fail on missing subcommand /
missing sock / unchanged file content.

**Naming note:** the live-runner grouping is `live-session/` (not `session/`).
A directory named `session/` under this tree collides with the doctest
`session` package import and breaks compilation.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

// SecondaryPack describes an extra host-built sealed sandbox used as a
// --load-devbox / --runtime-load-devbox target (same idea as tests/sandbox).
type SecondaryPack struct {
	Output            string
	ExtraFiles        []string
	ExtraEnv          []string
	HomeLinked        bool
	RuntimeLoadDevbox []string
}

// Request drives notify-event CLI leaves and/or a live sealed session.
type Request struct {
	// --- mode ---
	// HelpNotifyEvent runs `kool sandbox notify-event --help`.
	HelpNotifyEvent bool
	// RunNotifyEvent runs `kool sandbox notify-event` with Event* fields.
	RunNotifyEvent bool
	// LiveSession builds primary (+ SecondaryPacks) and starts a long guest.
	LiveSession bool

	// --- notify-event flags ---
	EventType string // --type (default leaf: devbox.updated)
	EventPath string // --path ABS
	// EventRoot is --root DIR (sandbox parent containing events/). Empty →
	// WorkingDir/kool-sandbox-root for notify leaves that create fixtures there.
	EventRoot string
	EventRootSet bool
	DryRun    bool

	// EnsureEventsDir creates EventRoot/events (0700) before notify when set.
	EnsureEventsDir bool
	// FakeSockNames are basenames created as empty files under events/ (for
	// dry-run listing without live listeners). Prefer MockListener for delivers.
	FakeSockNames []string
	// MockListener starts a real unix listener at events/<name> and captures
	// inbound payloads (delivers leaf).
	MockListener bool
	MockSockName string // default "sess-mock.sock"

	// --- sealed build / live session (host GOOS/GOARCH) ---
	Output            string
	OutputSet         bool
	ExtraFiles        []string
	ExtraEnv          []string
	RuntimeLoadDevbox []string
	SealedLoadDevbox  []string
	SealedDoubleDash  bool
	SecondaryPacks    []SecondaryPack
	// SandboxRootParent is KOOL_SANDBOX_ROOT for the sealed process.
	SandboxRootParent string
	// GuestShell is the shell body for the long-lived guest (sh -c …).
	// Empty → default wait-for-.guest-stop loop that writes .guest-ready.
	GuestShell string
	// GuestSleep is used when GuestShell empty and UseSleepGuest true.
	UseSleepGuest bool
	GuestSleep    time.Duration

	// After live start:
	// Poll for sock under events/ for up to SockWait.
	SockWait time.Duration
	// StopGuest writes .guest-stop into the session root (clean exit).
	StopGuest bool
	// WaitGuestExit waits for the sealed process to finish after stop/sleep.
	WaitGuestExit bool

	// Hot-reload / filter path:
	// RebuildLoadAfterStart rebuilds SecondaryPacks[0] with RebuildLoadFiles
	// (WorkingDir-relative --file LOCAL=SANDBOX_REL) after sock is up, then
	// runs notify-event for NotifyLoadPath (default: SecondaryPaths[0]).
	RebuildLoadAfterStart bool
	RebuildLoadFiles      []string // e.g. "load-new.txt=reload-me.txt"
	// NotifyAfterStart runs notify-event once sock is up (or after rebuild).
	NotifyAfterStart bool
	// NotifyLoadPath overrides --path for the post-start notify (absolute).
	// Empty → SecondaryPaths[0] when available; else EventPath.
	NotifyLoadPath string
	// NotifyEventType overrides --type for post-start notify (default devbox.updated).
	NotifyEventType string

	// ReadSessionRel is a sandbox-relative path under the live session root
	// to snapshot after notify (and optional settle).
	ReadSessionRel string
	// SettleAfterNotify waits this long after notify before reading session file.
	SettleAfterNotify time.Duration

	// SnapshotFileBeforeNotify reads ReadSessionRel once before notify (filter leaf).
	SnapshotFileBeforeNotify bool

	// WorkingDir is the process cwd isolation root (root Setup).
	WorkingDir string
	// ProcessTimeout bounds kool / sealed / notify subprocesses.
	ProcessTimeout time.Duration
}

// Response is CLI + live-session capture after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int

	// notify-event (primary mode or post-start notify)
	NotifyStdout   string
	NotifyStderr   string
	NotifyExitCode int
	NotifyRan      bool

	// mock listener
	DeliveredRaw   []string
	DeliveredCount int
	// FirstDelivered is parsed first JSON object when valid.
	FirstDelivered map[string]interface{}

	// build / live
	OutputPath     string
	OutputExists   bool
	OutputSize     int64
	SecondaryPaths []string
	SandboxRootParent string
	EventsDir      string
	SessionID      string
	SockPath       string
	SockExistsAfterStart bool
	SockMode       os.FileMode // 0 if unknown
	EventsDirMode  os.FileMode
	SessionRoot    string
	GuestReady     bool
	// SessionFileContent is the content of ReadSessionRel after settle.
	SessionFileContent string
	SessionFileExists  bool
	// SessionFileBefore is content when SnapshotFileBeforeNotify.
	SessionFileBefore       string
	SessionFileBeforeExists bool

	// after guest exit
	GuestExitCode       int
	GuestExited         bool
	SockExistsAfterExit bool
	MaterializeRemaining []string
}

func moduleRoot(doctestRoot string) string {
	return filepath.Clean(filepath.Join(doctestRoot, "..", ".."))
}

func sessionCacheDir(sessionID string) string {
	return filepath.Join(os.TempDir(), "kool-sandbox-events-doctest-"+sessionID)
}

func withFileLock(t *testing.T, lockPath string, fn func() error) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func ensureKoolBinary(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	cacheDir := sessionCacheDir(d.DOCTEST_SESSION_ID)
	lock := filepath.Join(cacheDir, "build.lock")
	ready := filepath.Join(cacheDir, "binaries.ready")
	bin := filepath.Join(cacheDir, "kool")
	modRoot := moduleRoot(d.DOCTEST_ROOT)
	err := withFileLock(t, lock, func() error {
		if st, err := os.Stat(ready); err == nil && !st.IsDir() {
			if st2, err2 := os.Stat(bin); err2 == nil && !st2.IsDir() {
				return nil
			}
		}
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return err
		}
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("go build kool: %w\n%s", err, out)
		}
		return os.WriteFile(ready, []byte("ok\n"), 0644)
	})
	if err != nil {
		return "", err
	}
	return bin, nil
}

func resolvePath(workingDir, p string) string {
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(workingDir, p)
}

func writeLocalFile(t *testing.T, workingDir, rel, content string) (string, error) {
	t.Helper()
	p := filepath.Join(workingDir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		return "", err
	}
	return p, nil
}

func runKool(t *testing.T, koolBin, workingDir string, timeout time.Duration, args []string, env []string) (stdout, stderr string, exitCode int, runErr error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 3 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, koolBin, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	if len(env) > 0 {
		cmd.Env = env
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout, stderr, -1, fmt.Errorf("kool exceeded process timeout %v; stderr=%q", timeout, stderr)
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout, stderr, exitErr.ExitCode(), nil
		}
		return stdout, stderr, -1, fmt.Errorf("run kool: %w", err)
	}
	return stdout, stderr, 0, nil
}

func buildSecondaryArgs(sp SecondaryPack) []string {
	args := []string{"sandbox", "build", "-o", sp.Output}
	for _, f := range sp.ExtraFiles {
		args = append(args, "--file", f)
	}
	for _, e := range sp.ExtraEnv {
		args = append(args, "--env", e)
	}
	if sp.HomeLinked {
		args = append(args, "--home-linked")
	}
	for _, p := range sp.RuntimeLoadDevbox {
		args = append(args, "--runtime-load-devbox", p)
	}
	return args
}

func buildSecondaryPacks(t *testing.T, koolBin, workingDir string, timeout time.Duration, packs []SecondaryPack) ([]string, error) {
	t.Helper()
	paths := make([]string, 0, len(packs))
	for i, sp := range packs {
		if sp.Output == "" {
			return nil, fmt.Errorf("SecondaryPacks[%d]: empty Output", i)
		}
		args := buildSecondaryArgs(sp)
		stdout, stderr, code, runErr := runKool(t, koolBin, workingDir, timeout, args, nil)
		if runErr != nil {
			return nil, fmt.Errorf("SecondaryPacks[%d] build: %w", i, runErr)
		}
		if code != 0 {
			return nil, fmt.Errorf("SecondaryPacks[%d] build exit=%d stdout=%q stderr=%q", i, code, stdout, stderr)
		}
		outPath := resolvePath(workingDir, sp.Output)
		st, err := os.Stat(outPath)
		if err != nil || st.IsDir() || st.Size() <= 0 {
			return nil, fmt.Errorf("SecondaryPacks[%d]: missing binary at %q", i, outPath)
		}
		paths = append(paths, outPath)
	}
	return paths, nil
}

func buildPrimaryArgs(req *Request) []string {
	args := []string{"sandbox", "build", "-o", req.Output}
	for _, f := range req.ExtraFiles {
		args = append(args, "--file", f)
	}
	for _, e := range req.ExtraEnv {
		args = append(args, "--env", e)
	}
	for _, p := range req.RuntimeLoadDevbox {
		args = append(args, "--runtime-load-devbox", p)
	}
	return args
}

func sealedRunArgs(req *Request) []string {
	var args []string
	for _, p := range req.SealedLoadDevbox {
		args = append(args, "--load-devbox", p)
	}
	if req.SealedDoubleDash {
		args = append(args, "--")
	}
	guest := req.GuestShell
	if guest == "" {
		if req.UseSleepGuest {
			sec := int(req.GuestSleep / time.Second)
			if sec < 1 {
				sec = 5
			}
			guest = fmt.Sprintf("sleep %d", sec)
		} else {
			// Write ready marker; wait for stop marker (host writes .guest-stop).
			guest = `echo ready > .guest-ready; while [ ! -f .guest-stop ]; do sleep 0.05; done`
		}
	}
	args = append(args, "sh", "-c", guest)
	return args
}

func notifyEventArgs(req *Request, typeOverride, pathOverride, root string, dryRun bool) []string {
	args := []string{"sandbox", "notify-event"}
	typ := typeOverride
	if typ == "" {
		typ = req.EventType
	}
	if typ == "" {
		typ = "devbox.updated"
	}
	path := pathOverride
	if path == "" {
		path = req.EventPath
	}
	args = append(args, "--type", typ)
	if path != "" {
		args = append(args, "--path", path)
	}
	if root != "" {
		args = append(args, "--root", root)
	}
	if dryRun || req.DryRun {
		args = append(args, "--dry-run")
	}
	return args
}

type mockSockServer struct {
	mu       sync.Mutex
	payloads []string
	ln       net.Listener
}

func (m *mockSockServer) add(p string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.payloads = append(m.payloads, p)
}

func (m *mockSockServer) list() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.payloads))
	copy(out, m.payloads)
	return out
}

func startMockSock(t *testing.T, sockPath string) (*mockSockServer, error) {
	t.Helper()
	_ = os.Remove(sockPath)
	if err := os.MkdirAll(filepath.Dir(sockPath), 0700); err != nil {
		return nil, err
	}
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, err
	}
	_ = os.Chmod(sockPath, 0600)
	m := &mockSockServer{ln: ln}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
				b, err := io.ReadAll(c)
				if err != nil && len(b) == 0 {
					return
				}
				// Accept one JSON object per connection (optional trailing newline).
				raw := strings.TrimSpace(string(b))
				if raw != "" {
					m.add(raw)
				}
			}(conn)
		}
	}()
	t.Cleanup(func() {
		ln.Close()
		_ = os.Remove(sockPath)
	})
	return m, nil
}

func eventsDir(root string) string {
	return filepath.Join(root, "events")
}

func listSocks(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sock") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func waitForSock(events string, timeout time.Duration) (sockPath string, ok bool) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		names, err := listSocks(events)
		if err == nil && len(names) > 0 {
			return filepath.Join(events, names[0]), true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return "", false
}

func findSessionRoot(parent, sessionID string) string {
	if sessionID != "" {
		p := filepath.Join(parent, sessionID)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	// Fallback: first non-events directory under parent.
	ents, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}
	for _, e := range ents {
		if !e.IsDir() || e.Name() == "events" {
			continue
		}
		return filepath.Join(parent, e.Name())
	}
	return ""
}

func sessionIDFromSock(sockPath string) string {
	base := filepath.Base(sockPath)
	return strings.TrimSuffix(base, ".sock")
}

func readFileIfExists(path string) (string, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return string(b), true, nil
}

func listDirNames(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(ents))
	for _, e := range ents {
		names = append(names, e.Name())
	}
	return names, nil
}

// Run executes help/notify-event and/or live sealed session scenarios.
// Author-named d *session.Doctest is required in one-process mode.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	koolBin, err := ensureKoolBinary(t, d)
	if err != nil {
		return nil, err
	}
	timeout := req.ProcessTimeout
	if timeout <= 0 {
		timeout = 3 * time.Minute
	}
	resp := &Response{}

	// Resolve sandbox root parent used by notify + live session.
	// Prefer a short absolute parent so parent/events/*.sock fits AF_UNIX
	// (macOS sun_path ~104). Long paths under t.TempDir() cause bind EINVAL.
	parent := req.SandboxRootParent
	if parent == "" {
		parent = req.EventRoot
	}
	if parent == "" {
		dir, mkErr := os.MkdirTemp("/tmp", "ksb-")
		if mkErr != nil {
			return nil, mkErr
		}
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		parent = dir
	} else {
		parent = resolvePath(req.WorkingDir, parent)
	}
	resp.SandboxRootParent = parent
	resp.EventsDir = eventsDir(parent)

	// --- help ---
	if req.HelpNotifyEvent {
		args := []string{"sandbox", "notify-event", "--help"}
		stdout, stderr, code, runErr := runKool(t, koolBin, req.WorkingDir, timeout, args, nil)
		resp.Stdout = stdout
		resp.Stderr = stderr
		resp.ExitCode = code
		return resp, runErr
	}

	// Prepare events dir fixtures for pure notify leaves.
	if req.EnsureEventsDir || len(req.FakeSockNames) > 0 || req.MockListener {
		if err := os.MkdirAll(resp.EventsDir, 0700); err != nil {
			return resp, err
		}
	}
	for _, name := range req.FakeSockNames {
		p := filepath.Join(resp.EventsDir, name)
		if err := os.WriteFile(p, []byte{}, 0600); err != nil {
			return resp, err
		}
	}
	var mock *mockSockServer
	if req.MockListener {
		name := req.MockSockName
		if name == "" {
			name = "sess-mock.sock"
		}
		sockPath := filepath.Join(resp.EventsDir, name)
		mock, err = startMockSock(t, sockPath)
		if err != nil {
			return resp, fmt.Errorf("mock sock: %w", err)
		}
		resp.SockPath = sockPath
	}

	// --- pure notify-event (no live session) ---
	if req.RunNotifyEvent && !req.LiveSession {
		root := parent
		if req.EventRootSet || req.EventRoot != "" {
			root = resolvePath(req.WorkingDir, req.EventRoot)
			if root == "" {
				root = parent
			}
		}
		args := notifyEventArgs(req, req.EventType, req.EventPath, root, req.DryRun)
		env := append(os.Environ(), "KOOL_SANDBOX_ROOT="+root)
		stdout, stderr, code, runErr := runKool(t, koolBin, req.WorkingDir, timeout, args, env)
		resp.Stdout = stdout
		resp.Stderr = stderr
		resp.ExitCode = code
		resp.NotifyStdout = stdout
		resp.NotifyStderr = stderr
		resp.NotifyExitCode = code
		resp.NotifyRan = true
		if mock != nil {
			// Brief settle for async accept.
			time.Sleep(100 * time.Millisecond)
			resp.DeliveredRaw = mock.list()
			resp.DeliveredCount = len(resp.DeliveredRaw)
			if resp.DeliveredCount > 0 {
				var m map[string]interface{}
				if json.Unmarshal([]byte(resp.DeliveredRaw[0]), &m) == nil {
					resp.FirstDelivered = m
				}
			}
		}
		return resp, runErr
	}

	// --- live session ---
	if !req.LiveSession {
		return resp, fmt.Errorf("Run: no mode set (HelpNotifyEvent / RunNotifyEvent / LiveSession)")
	}

	if err := os.MkdirAll(parent, 0755); err != nil {
		return resp, err
	}

	if len(req.SecondaryPacks) > 0 {
		secPaths, secErr := buildSecondaryPacks(t, koolBin, req.WorkingDir, timeout, req.SecondaryPacks)
		if secErr != nil {
			return resp, secErr
		}
		resp.SecondaryPaths = secPaths
	}

	outName := req.Output
	if outName == "" {
		outName = "sandbox.bin"
	}
	req.Output = outName
	buildArgs := buildPrimaryArgs(req)
	stdout, stderr, code, runErr := runKool(t, koolBin, req.WorkingDir, timeout, buildArgs, nil)
	resp.Stdout = stdout
	resp.Stderr = stderr
	resp.ExitCode = code
	if runErr != nil {
		return resp, runErr
	}
	outPath := resolvePath(req.WorkingDir, outName)
	resp.OutputPath = outPath
	if st, err := os.Stat(outPath); err == nil && !st.IsDir() {
		resp.OutputExists = true
		resp.OutputSize = st.Size()
	}
	if code != 0 || !resp.OutputExists {
		return resp, nil
	}

	// Start sealed binary in background.
	sArgs := sealedRunArgs(req)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// cancel on cleanup; also used if WaitGuestExit.
	cmd := exec.CommandContext(ctx, outPath, sArgs...)
	cmd.Dir = req.WorkingDir
	cmd.Env = append(os.Environ(), "KOOL_SANDBOX_ROOT="+parent)
	var guestOut, guestErr bytes.Buffer
	cmd.Stdout = &guestOut
	cmd.Stderr = &guestErr
	if err := cmd.Start(); err != nil {
		cancel()
		return resp, fmt.Errorf("start sealed: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	// Single-consumer wait so Cleanup and WaitGuestExit never double-read done.
	var guestWaitMu sync.Mutex
	var guestWaitFinished bool
	var guestWaitErr error
	waitGuest := func() error {
		guestWaitMu.Lock()
		defer guestWaitMu.Unlock()
		if guestWaitFinished {
			return guestWaitErr
		}
		guestWaitErr = <-done
		guestWaitFinished = true
		return guestWaitErr
	}

	// Always try to stop guest on exit.
	t.Cleanup(func() {
		// Best-effort: write stop marker if session known.
		if resp.SessionRoot != "" {
			_ = os.WriteFile(filepath.Join(resp.SessionRoot, ".guest-stop"), []byte("1\n"), 0644)
		}
		finished := make(chan struct{})
		go func() {
			_ = waitGuest()
			close(finished)
		}()
		select {
		case <-finished:
		case <-time.After(2 * time.Second):
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			<-finished
		}
		cancel()
	})

	sockWait := req.SockWait
	if sockWait <= 0 {
		sockWait = 12 * time.Second
	}
	// Also poll for .guest-ready under any session dir (guest may start before sock).
	deadline := time.Now().Add(sockWait)
	for time.Now().Before(deadline) {
		if sp, ok := waitForSock(resp.EventsDir, 100*time.Millisecond); ok {
			resp.SockPath = sp
			resp.SockExistsAfterStart = true
			resp.SessionID = sessionIDFromSock(sp)
			break
		}
		// Discover session via materialize children even without sock (RED path).
		if resp.SessionRoot == "" {
			resp.SessionRoot = findSessionRoot(parent, "")
			if resp.SessionRoot != "" {
				resp.SessionID = filepath.Base(resp.SessionRoot)
				if _, err := os.Stat(filepath.Join(resp.SessionRoot, ".guest-ready")); err == nil {
					resp.GuestReady = true
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	if resp.SockExistsAfterStart {
		if st, err := os.Stat(resp.SockPath); err == nil {
			resp.SockMode = st.Mode().Perm()
		}
		if st, err := os.Stat(resp.EventsDir); err == nil {
			resp.EventsDirMode = st.Mode().Perm()
		}
		// Prefer session id derived from sock basename.
		if resp.SessionID != "" {
			if p := findSessionRoot(parent, resp.SessionID); p != "" {
				resp.SessionRoot = p
			}
		}
		if resp.SessionRoot == "" {
			resp.SessionRoot = findSessionRoot(parent, "")
		}
	}
	if resp.SessionRoot != "" {
		if _, err := os.Stat(filepath.Join(resp.SessionRoot, ".guest-ready")); err == nil {
			resp.GuestReady = true
		}
	}

	if req.SnapshotFileBeforeNotify && req.ReadSessionRel != "" && resp.SessionRoot != "" {
		content, ok, rerr := readFileIfExists(filepath.Join(resp.SessionRoot, filepath.FromSlash(req.ReadSessionRel)))
		if rerr != nil {
			return resp, rerr
		}
		resp.SessionFileBefore = content
		resp.SessionFileBeforeExists = ok
	}

	if req.RebuildLoadAfterStart && len(req.SecondaryPacks) > 0 {
		sp := req.SecondaryPacks[0]
		if len(req.RebuildLoadFiles) > 0 {
			sp.ExtraFiles = req.RebuildLoadFiles
		}
		// Rebuild in place (same -o path).
		paths, berr := buildSecondaryPacks(t, koolBin, req.WorkingDir, timeout, []SecondaryPack{sp})
		if berr != nil {
			// Record error path but still try to stop guest.
			if req.StopGuest && resp.SessionRoot != "" {
				_ = os.WriteFile(filepath.Join(resp.SessionRoot, ".guest-stop"), []byte("1\n"), 0644)
			}
			return resp, berr
		}
		if len(paths) > 0 {
			if len(resp.SecondaryPaths) == 0 {
				resp.SecondaryPaths = paths
			} else {
				resp.SecondaryPaths[0] = paths[0]
			}
		}
	}

	if req.NotifyAfterStart {
		nPath := req.NotifyLoadPath
		if nPath == "" {
			if len(resp.SecondaryPaths) > 0 {
				nPath = resp.SecondaryPaths[0]
			} else {
				nPath = req.EventPath
			}
		}
		nType := req.NotifyEventType
		if nType == "" {
			nType = "devbox.updated"
		}
		args := notifyEventArgs(req, nType, nPath, parent, false)
		env := append(os.Environ(), "KOOL_SANDBOX_ROOT="+parent)
		nOut, nErr, nCode, nRunErr := runKool(t, koolBin, req.WorkingDir, timeout, args, env)
		resp.NotifyStdout = nOut
		resp.NotifyStderr = nErr
		resp.NotifyExitCode = nCode
		resp.NotifyRan = true
		if nRunErr != nil {
			if req.StopGuest && resp.SessionRoot != "" {
				_ = os.WriteFile(filepath.Join(resp.SessionRoot, ".guest-stop"), []byte("1\n"), 0644)
			}
			return resp, nRunErr
		}
		settle := req.SettleAfterNotify
		if settle <= 0 {
			settle = 300 * time.Millisecond
		}
		time.Sleep(settle)
	}

	if req.ReadSessionRel != "" && resp.SessionRoot != "" {
		content, ok, rerr := readFileIfExists(filepath.Join(resp.SessionRoot, filepath.FromSlash(req.ReadSessionRel)))
		if rerr != nil {
			return resp, rerr
		}
		resp.SessionFileContent = content
		resp.SessionFileExists = ok
	}

	if req.StopGuest && resp.SessionRoot != "" {
		_ = os.WriteFile(filepath.Join(resp.SessionRoot, ".guest-stop"), []byte("1\n"), 0644)
	}

	if req.WaitGuestExit || req.StopGuest || req.UseSleepGuest {
		finished := make(chan error, 1)
		go func() { finished <- waitGuest() }()
		select {
		case err := <-finished:
			resp.GuestExited = true
			if err == nil {
				resp.GuestExitCode = 0
			} else if ee, ok := err.(*exec.ExitError); ok {
				resp.GuestExitCode = ee.ExitCode()
			} else {
				resp.GuestExitCode = -1
			}
		case <-time.After(15 * time.Second):
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			err := <-finished
			resp.GuestExited = true
			if ee, ok := err.(*exec.ExitError); ok {
				resp.GuestExitCode = ee.ExitCode()
			} else {
				resp.GuestExitCode = -1
			}
		}
		// Snapshot sock after exit.
		if resp.SockPath != "" {
			_, err := os.Stat(resp.SockPath)
			resp.SockExistsAfterExit = err == nil
		} else {
			// Any remaining socks?
			names, _ := listSocks(resp.EventsDir)
			resp.SockExistsAfterExit = len(names) > 0
		}
		names, _ := listDirNames(parent)
		resp.MaterializeRemaining = names
	}

	// Append guest streams for debugging on RED.
	if guestOut.Len() > 0 {
		resp.Stdout = resp.Stdout + "\n[guest-stdout]\n" + guestOut.String()
	}
	if guestErr.Len() > 0 {
		resp.Stderr = resp.Stderr + "\n[guest-stderr]\n" + guestErr.String()
	}

	return resp, nil
}
```
