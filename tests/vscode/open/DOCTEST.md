# kool vscode open

`kool vscode open <dir>` validates a local directory, runs VS Code precheck (`code`
CLI + extension listed), opens via IPC to `xhd2015.open-in-new-window`, and falls
back to `vscode://.../open?path=<encoded>` when IPC is unreachable. By default the
extension opens in a new window (focusing an existing window when the dir is already
open). `--replace` reuses the current window instead. `--ipc-only` skips URI/OS
fallback (probe mode). `--json` emits one JSON object on stdout and suppresses human
IPC fallback hints on stderr.

## Version

0.0.2

## DSN (Domain Specific Notion)

### Participants
- **kool CLI** — `vscode.go` subcommand `open`; orchestrates validate → precheck →
  IPC → URI fallback.
- **`ValidateDirPath`** — resolves path against cwd, checks exists and is directory;
  returns normalized absolute path (no `.git` requirement).
- **`EnsureCodeCLI`** — verifies `code` is on PATH (injectable command for tests).
- **`EnsureExtensionListed`** — runs `code --list-extensions`, requires
  `xhd2015.open-in-new-window`.
- **`BuildOpenURI`** — constructs `vscode://xhd2015.open-in-new-window/open?path=...`
  with URL-encoded absolute path; appends `&replace=true` when `--replace` is set.
- **IPC client** — JSON-lines over Unix socket `~/.kool/xhd2015.open-in-new-window.sock`;
  op `open` with absolute path; optional `replace: true`; 3 retries × 100ms.
- **OS opener** — `open` (darwin), `xdg-open` (linux), `cmd /c start` (windows);
  injectable `execCommand` for tests; used only on IPC failure with stderr hint.
- **VS Code extension** — receives IPC `open` or URI handler; default uses
  `forceNewWindow: true` (new window, focus if dir already open); `--replace` uses
  `forceNewWindow: false` (reuse current window).

### Behaviors
- **Validation failure** — missing arg, nonexistent path, or not a directory → stderr
  error, non-zero exit, no precheck/IPC/exec.
- **Precheck failure** — `code` missing or extension not listed → stderr with hint,
  non-zero exit, no IPC/exec.
- **IPC success** — extension acknowledges `{"ok":true}`; no OS opener invoked.
- **IPC retry** — transient connect failures retried; success on later attempt.
- **IPC failure + fallback** — stderr note about IPC unreachable; OS opener invoked
  with `vscode://.../open?path=...` (and `&replace=true` when flagged).
- **URI building** — absolute/relative/trailing-slash paths normalize to encoded URI;
  `replace` query param omitted by default, present when `--replace`.
- **`--replace` flag** — boolean CLI flag on `open` only; propagates to IPC JSON and
  URI fallback; does not affect validation or precheck.
- **`--ipc-only` flag** — after validate + precheck, IPC only; IPC failure returns
  error with non-zero exit; no stderr URI-fallback hint and no OS opener.
- **`--json` flag** — prints a single JSON object on stdout (`ipc_handled`, `path`,
  optional `error`, `fallback`, `ok`); suppresses human IPC-unreachable stderr hint.
- **CLI subprocess IPC override** — `KOOL_VSCODE_IPC_SOCKET` env points the CLI at a
  test mock socket (same path as in-process `SetIPC_SOCKETPathForTest`).

## Decision Tree

```
open/
├── validation/                 [path invalid — blocks pipeline]
│   ├── missing-arg/            → CLI usage error; no open
│   ├── nonexistent-path/       → error before precheck
│   └── not-directory/          → error before precheck
├── precheck/                   [environment invalid — blocks IPC/exec]
│   ├── no-code-cli/            → error mentions code / PATH
│   └── extension-not-listed/   → error mentions extension id + install hint
├── cli/                        [CLI flags + subprocess]
│   ├── replace-flag/           → kool vscode open --replace <dir> succeeds
│   ├── ipc-only-json-success/  → --ipc-only --json, mock IPC → JSON success, exit 0
│   └── ipc-only-json-failure/  → --ipc-only --json, IPC fail → JSON failure, exit 1
├── uri/                        [URI construction — no side effects]
│   ├── absolute-path/          → correct vscode:// URI (default)
│   ├── relative-path/          → cwd-resolved absolute in URI
│   ├── trailing-slash/         → normalized path in URI
│   ├── default-no-replace/     → URI omits replace= query param
│   └── replace-query/          → URI includes replace=true
├── ipc/                        [IPC delivery — primary path]
│   ├── success-no-fallback/    → IPC called; no replace; OS open NOT called
│   ├── retry-then-success/     → 1st connect fails; 2nd succeeds
│   ├── fail-then-fallback/     → stderr hint + open with vscode:// URI
│   ├── default-new-window/     → IPC JSON omits replace field
│   ├── replace-flag/           → IPC JSON has replace:true
│   ├── ipc-only-success/       → --ipc-only, IPC ok → no OS opener
│   └── ipc-only-no-fallback/   → --ipc-only, IPC fail → error; no hint; no exec
├── json/                       [--json machine-readable output]
│   └── json-with-uri-fallback/ → --json, IPC fail + URI ok → JSON fallback:uri
└── exec/                       [URI fallback delivery]
    └── fallback-invokes-open/  → mock exec; opener called with URI after IPC fail
```

## Test Index

| # | Path | Description |
|---|------|-------------|
| 1 | `validation/missing-arg/` | No path argument shows usage error |
| 2 | `validation/nonexistent-path/` | Nonexistent path fails before precheck |
| 3 | `validation/not-directory/` | File path fails before precheck |
| 4 | `precheck/no-code-cli/` | Missing `code` CLI blocks open |
| 5 | `precheck/extension-not-listed/` | Unlisted extension blocks open |
| 6 | `cli/replace-flag/` | CLI accepts `--replace` and exits 0 |
| 7 | `uri/absolute-path/` | Absolute path produces correct URI (default) |
| 8 | `uri/relative-path/` | Relative path resolved in URI |
| 9 | `uri/trailing-slash/` | Trailing slash stripped in URI |
| 10 | `uri/default-no-replace/` | Default URI omits `replace=` query |
| 11 | `uri/replace-query/` | `--replace` adds `replace=true` to URI |
| 12 | `ipc/success-no-fallback/` | IPC success; no replace; OS opener not called |
| 13 | `ipc/retry-then-success/` | First IPC connect fails; retry succeeds |
| 14 | `ipc/fail-then-fallback/` | IPC exhausted; stderr hint + URI fallback |
| 15 | `ipc/default-new-window/` | IPC JSON omits `replace` field |
| 16 | `ipc/replace-flag/` | IPC JSON includes `replace:true` |
| 17 | `exec/fallback-invokes-open/` | Full orchestration falls back to OS opener |
| 18 | `ipc/ipc-only-success/` | `--ipc-only` with IPC ok; OS opener not called |
| 19 | `ipc/ipc-only-no-fallback/` | `--ipc-only` IPC fail; no URI hint or exec |
| 20 | `json/json-with-uri-fallback/` | `--json` reports URI fallback when IPC fails |
| 21 | `cli/ipc-only-json-success/` | `--ipc-only --json` probe success on stdout |
| 22 | `cli/ipc-only-json-failure/` | `--ipc-only --json` probe failure on stdout |

## How to Run

```sh
doctest vet ./tests/vscode/open
doctest test ./tests/vscode/open
go test ./vscodegit/...
```

```go
import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	vscodegit "github.com/xhd2015/kool/vscodegit"
)

const koolVscodeIPCSocketEnv = "KOOL_VSCODE_IPC_SOCKET"

type Request struct {
	Phase           string
	DirPath         string
	Replace         bool
	IpcOnly         bool
	Json            bool
	WorkingDir      string
	GoOS            string
	CodeCommand     string
	CodeInPath      bool
	IPCSocketPath   string
	IPCFailConnects int
	IPCAlwaysFail   bool
}

type openJSONPayload struct {
	IPCHandled bool   `json:"ipc_handled"`
	Path       string `json:"path"`
	Error      string `json:"error,omitempty"`
	Fallback   string `json:"fallback,omitempty"`
	OK         *bool  `json:"ok,omitempty"`
}

type Response struct {
	NormalizedPath string
	VSCodeURI      string
	Stdout         string
	Stderr         string
	ExitCode       int
	ValidateErr    string
	PrecheckErr    string
	ExecCalled     bool
	ExecCommand    string
	ExecArgs       []string
	IPCCalled      bool
	IPCAttempts    int
	IPCOp          string
	IPCPath        string
	IPCReplaceSet  bool
	IPCReplace     bool
	StderrHint     bool
	OpenJSON       *openJSONPayload
}

type ipcServerState struct {
	mu           sync.Mutex
	requests     []map[string]interface{}
	connectCount int
	failFirst    int
	alwaysReject bool
}

func resolveKoolBinary() (string, error) {
	moduleRoot := filepath.Join(DOCTEST_ROOT, "..", "..", "..")
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
	return "", fmt.Errorf("kool binary not found in PATH or %s; build with: go build -o kool .", moduleRoot)
}

func configurePrecheckForCLI(t *testing.T, req *Request, cmd *exec.Cmd) {
	env := os.Environ()
	if req.CodeCommand != "" {
		dir := filepath.Dir(req.CodeCommand)
		env = append(env, "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	if !req.CodeInPath {
		env = append(env, "PATH=/empty")
	}
	if req.IPCSocketPath != "" {
		env = append(env, koolVscodeIPCSocketEnv+"="+req.IPCSocketPath)
	}
	cmd.Env = env
}

func parseOpenJSON(t *testing.T, stdout string) *openJSONPayload {
	t.Helper()
	line := strings.TrimSpace(stdout)
	if line == "" {
		t.Fatal("expected JSON on stdout, got empty")
	}
	var payload openJSONPayload
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		t.Fatalf("parse open JSON: %v\nstdout: %s", err, stdout)
	}
	return &payload
}

func attachOpenJSON(t *testing.T, resp *Response) {
	t.Helper()
	if strings.TrimSpace(resp.Stdout) == "" {
		return
	}
	resp.OpenJSON = parseOpenJSON(t, resp.Stdout)
}

func installPrecheckHooks(t *testing.T, req *Request) func() {
	if req.CodeCommand != "" {
		vscodegit.SetCodeCommandForTest(req.CodeCommand)
		t.Cleanup(func() { vscodegit.SetCodeCommandForTest("") })
	}
	if !req.CodeInPath {
		vscodegit.SetCodeCommandForTest("")
		t.Cleanup(func() { vscodegit.SetCodeCommandForTest("") })
	}
	return func() {}
}

func handleIPCConn(conn net.Conn, state *ipcServerState) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	var req map[string]interface{}
	_ = json.Unmarshal([]byte(strings.TrimSpace(line)), &req)
	state.mu.Lock()
	state.requests = append(state.requests, req)
	state.mu.Unlock()
	resp := map[string]interface{}{"ok": true}
	b, _ := json.Marshal(resp)
	_, _ = conn.Write(append(b, '\n'))
	time.Sleep(10 * time.Millisecond)
}

func startMockIPCServer(t *testing.T, socketPath string, failFirst int) *ipcServerState {
	t.Helper()
	_ = os.Remove(socketPath)
	_ = os.MkdirAll(filepath.Dir(socketPath), 0755)

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("mock IPC listen: %v", err)
	}

	state := &ipcServerState{failFirst: failFirst}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			state.mu.Lock()
			state.connectCount++
			reject := state.alwaysReject || state.failFirst > 0
			if state.failFirst > 0 {
				state.failFirst--
			}
			state.mu.Unlock()
			if reject {
				conn.Close()
				continue
			}
			go handleIPCConn(conn, state)
		}
	}()
	t.Cleanup(func() {
		ln.Close()
		_ = os.Remove(socketPath)
	})
	return state
}

func runCLI(t *testing.T, req *Request) (*Response, error) {
	koolBin, err := resolveKoolBinary()
	if err != nil {
		return nil, err
	}
	args := []string{"vscode", "open"}
	if req.Replace {
		args = append(args, "--replace")
	}
	if req.IpcOnly {
		args = append(args, "--ipc-only")
	}
	if req.Json {
		args = append(args, "--json")
	}
	if req.DirPath != "" {
		args = append(args, req.DirPath)
	}
	cmd := exec.Command(koolBin, args...)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}
	configurePrecheckForCLI(t, req, cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to run kool: %w", runErr)
		}
	}
	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

func runValidate(t *testing.T, req *Request) (*Response, error) {
	cwd := req.WorkingDir
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	normalized, err := vscodegit.ValidateDirPath(req.DirPath, cwd)
	resp := &Response{NormalizedPath: normalized}
	if err != nil {
		resp.ValidateErr = err.Error()
	}
	return resp, nil
}

func runBuildURI(t *testing.T, req *Request) (*Response, error) {
	cwd := req.WorkingDir
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	normalized, err := vscodegit.ValidateDirPath(req.DirPath, cwd)
	if err != nil {
		return &Response{ValidateErr: err.Error()}, nil
	}
	uri := vscodegit.BuildOpenURI(normalized, req.Replace)
	return &Response{
		NormalizedPath: normalized,
		VSCodeURI:      uri,
	}, nil
}

func runPrecheck(t *testing.T, req *Request) (*Response, error) {
	restore := installPrecheckHooks(t, req)
	defer restore()

	resp := &Response{}
	if err := vscodegit.EnsureCodeCLI(); err != nil {
		resp.PrecheckErr = err.Error()
		return resp, nil
	}
	if err := vscodegit.EnsureExtensionListed(); err != nil {
		resp.PrecheckErr = err.Error()
		return resp, nil
	}
	return resp, nil
}

func runOrchestrate(t *testing.T, req *Request) (*Response, error) {
	cwd := req.WorkingDir
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	var execCalled bool
	var execCommand string
	var execArgs []string
	vscodegit.SetExecCommandHook(func(name string, arg ...string) *exec.Cmd {
		execCalled = true
		execCommand = name
		execArgs = append([]string{}, arg...)
		return exec.Command("true")
	})
	t.Cleanup(func() { vscodegit.SetExecCommandHook(nil) })

	if req.GoOS != "" {
		vscodegit.SetGOOSForTest(req.GoOS)
		t.Cleanup(func() { vscodegit.SetGOOSForTest("") })
	}

	restorePrecheck := installPrecheckHooks(t, req)
	defer restorePrecheck()

	socketPath := req.IPCSocketPath
	if socketPath == "" {
		socketPath = filepath.Join(t.TempDir(), "ipc.sock")
	}
	vscodegit.SetIPC_SOCKETPathForTest(socketPath)
	t.Cleanup(func() { vscodegit.SetIPC_SOCKETPathForTest("") })

	var server *ipcServerState
	if !req.IPCAlwaysFail {
		server = startMockIPCServer(t, socketPath, req.IPCFailConnects)
	}

	var stderrBuf bytes.Buffer
	vscodegit.SetStderrWriterForTest(&stderrBuf)
	t.Cleanup(func() { vscodegit.SetStderrWriterForTest(nil) })

	var stdoutBuf bytes.Buffer
	emitJSON := req.Json || req.Phase == "json"
	if emitJSON {
		vscodegit.SetStdoutWriterForTest(&stdoutBuf)
		t.Cleanup(func() { vscodegit.SetStdoutWriterForTest(nil) })
	}

	opts := vscodegit.OpenOptions{
		Replace: req.Replace,
		IpcOnly: req.IpcOnly,
		Json:    emitJSON,
	}
	err := vscodegit.OpenDirOptions(req.DirPath, cwd, opts)
	resp := &Response{
		ExecCalled:  execCalled,
		ExecCommand: execCommand,
		ExecArgs:    execArgs,
		Stderr:      stderrBuf.String(),
		Stdout:      stdoutBuf.String(),
	}
	if err != nil {
		resp.ValidateErr = err.Error()
	}
	if execCalled && len(execArgs) > 0 {
		resp.VSCodeURI = execArgs[len(execArgs)-1]
	}
	if server != nil {
		server.mu.Lock()
		resp.IPCAttempts = server.connectCount
		resp.IPCCalled = len(server.requests) > 0
		if len(server.requests) > 0 {
			first := server.requests[0]
			if op, ok := first["op"].(string); ok {
				resp.IPCOp = op
			}
			if path, ok := first["path"].(string); ok {
				resp.IPCPath = path
			}
			if replace, ok := first["replace"]; ok {
				resp.IPCReplaceSet = true
				if b, ok := replace.(bool); ok {
					resp.IPCReplace = b
				}
			}
		}
		server.mu.Unlock()
	}
	resp.StderrHint = strings.Contains(resp.Stderr, "extension not reachable via IPC")
	if emitJSON {
		attachOpenJSON(t, resp)
	}
	return resp, nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Phase {
	case "cli":
		resp, err := runCLI(t, req)
		if err != nil {
			return nil, err
		}
		if req.Json {
			attachOpenJSON(t, resp)
		}
		return resp, nil
	case "validate":
		return runValidate(t, req)
	case "build-uri":
		return runBuildURI(t, req)
	case "precheck":
		return runPrecheck(t, req)
	case "ipc", "exec", "orchestrate", "json":
		return runOrchestrate(t, req)
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}
```