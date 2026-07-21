# kool vscode open-git-repo

`kool vscode open-git-repo <path>` validates a local git repository, builds a
`vscode://xhd2015.open-in-new-window/git-open?path=<encoded>` URI, and opens it
via the OS handler (`open` / `xdg-open` / `cmd /c start`).

## Version

0.0.2

## DSN (Domain Specific Notion)

### Participants
- **kool CLI** — `vscode.go` subcommand `open-git-repo`; orchestrates validate →
  build URI → exec open.
- **`validateGitRepoPath`** — resolves path against cwd, checks exists, directory,
  and `.git` (file or directory); returns normalized absolute path.
- **`buildGitOpenRepoURI`** — constructs `vscode://xhd2015.open-in-new-window/git-open?path=...`
  with URL-encoded absolute path.
- **OS opener** — `open` (darwin), `xdg-open` (linux), `cmd /c start` (windows);
  injectable `execCommand` for tests.
- **VS Code extension** — receives URI and calls `git.openRepository` (tested separately
  in `tests/git-open-cli/`).

- **`EnsureCodeCLI`** — verifies `code` is on PATH (injectable for tests).
- **`EnsureExtensionListed`** — runs `code --list-extensions`, requires extension id.
- **IPC client** — JSON-lines over Unix socket; op `git-open`; 3 retries × 100ms.
### Behaviors
- **Validation failure** — missing arg, nonexistent path, not a directory, or no `.git`
  → stderr error, non-zero exit, no precheck/IPC/exec.
- **Precheck failure** — `code` missing or extension not listed → stderr with hint, no IPC/exec.
- **IPC success** — extension acknowledges `{"ok":true}`; no OS opener invoked.
- **IPC failure + fallback** — stderr IPC hint; OS opener invoked with `vscode://.../git-open?path=...`.
- **URI building** — absolute/relative/trailing-slash/spaces paths normalize to encoded URI.
- **Exec (legacy phase)** — orchestration with IPC disabled falls back to OS opener with URI.

## Decision Tree

```
open-git-repo/
├── validation/
│   ├── missing-arg/        → stderr usage error; no open
│   ├── nonexistent-path/   → error before open
│   ├── not-directory/      → error before open
│   └── no-git/             → "not a git repository"
├── uri/
│   ├── absolute-path/      → correct vscode:// URI
│   ├── relative-path/      → cwd-resolved absolute in URI
│   ├── trailing-slash/     → normalized path in URI
│   ├── spaces-in-path/     → URL-encoded spaces
│   └── worktree-git-file/  → .git file worktree accepted
├── precheck/
│   ├── no-code-cli/        → error mentions code / PATH
│   └── extension-not-listed/ → error mentions extension id + install hint
├── ipc/
│   ├── success-no-fallback/ → IPC git-open; OS open NOT called
│   └── fail-then-fallback/  → stderr hint + open with git-open URI
└── exec/
    └── invokes-open/       → mock exec; opener called with URI (IPC disabled)
```

## Test Index

| # | Path | Description |
|---|------|-------------|
| 1 | `validation/missing-arg/` | No path argument shows usage error |
| 2 | `validation/nonexistent-path/` | Nonexistent path fails before open |
| 3 | `validation/not-directory/` | File path fails before open |
| 4 | `validation/no-git/` | Directory without `.git` fails |
| 5 | `uri/absolute-path/` | Absolute path produces correct URI |
| 6 | `uri/relative-path/` | Relative path resolved in URI |
| 7 | `uri/trailing-slash/` | Trailing slash stripped in URI |
| 8 | `uri/spaces-in-path/` | Spaces URL-encoded in URI |
| 9 | `uri/worktree-git-file/` | Worktree `.git` file accepted |
| 10 | `exec/invokes-open/` | OS opener invoked with built URI |
| 11 | `precheck/no-code-cli/` | Missing `code` CLI blocks open |
| 12 | `precheck/extension-not-listed/` | Unlisted extension blocks open |
| 13 | `ipc/success-no-fallback/` | IPC success; OS opener not called |
| 14 | `ipc/fail-then-fallback/` | IPC failure triggers URI fallback |

## How to Run

```sh
cd kool-vscode
doctest vet ./tests/vscode/open-git-repo
doctest test ./tests/vscode/open-git-repo
go test ./...
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

	"github.com/xhd2015/doctest/session"
	vscodegit "github.com/xhd2015/kool/vscodegit"
)

type Request struct {
	Phase           string
	RepoPath        string
	WorkingDir      string
	GoOS            string
	CodeCommand     string
	CodeInPath      bool
	IPCSocketPath   string
	IPCFailConnects int
	IPCAlwaysFail   bool
}

type Response struct {
	NormalizedPath string
	VSCodeURI      string
	Stdout         string
	Stderr         string
	ExitCode       int
	ExecCalled     bool
	ExecCommand    string
	ExecArgs       []string
	ValidateErr    string
	PrecheckErr    string
	IPCCalled      bool
	IPCAttempts    int
	IPCOp          string
	IPCPath        string
	StderrHint     bool
}

type gitIPCServerState struct {
	mu           sync.Mutex
	requests     []map[string]string
	connectCount int
	failFirst    int
}

func resolveKoolBinary(d *session.Doctest) (string, error) {
	moduleRoot := filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", "..", ".."))
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
	if req.CodeCommand != "" {
		dir := filepath.Dir(req.CodeCommand)
		cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	if !req.CodeInPath {
		cmd.Env = append(cmd.Env, "PATH=/empty")
	}
}

func installPrecheckHooks(t *testing.T, req *Request) {
	if req.CodeCommand != "" {
		vscodegit.SetCodeCommandForTest(req.CodeCommand)
		t.Cleanup(func() { vscodegit.SetCodeCommandForTest("") })
	}
	if !req.CodeInPath {
		vscodegit.SetCodeCommandForTest("")
		t.Cleanup(func() { vscodegit.SetCodeCommandForTest("") })
	} else if req.CodeCommand == "" && req.Phase != "validate" && req.Phase != "build-uri" && req.Phase != "cli" {
		binDir := filepath.Join(req.WorkingDir, "bin")
		_ = os.MkdirAll(binDir, 0755)
		script := filepath.Join(binDir, "code")
		body := "#!/bin/sh\ncase \"$1\" in\n--list-extensions)\n  echo 'xhd2015.open-in-new-window'\n  ;;\nesac\n"
		_ = os.WriteFile(script, []byte(body), 0755)
		vscodegit.SetCodeCommandForTest(script)
		t.Cleanup(func() { vscodegit.SetCodeCommandForTest("") })
	}
}

func handleGitIPCConn(conn net.Conn, state *gitIPCServerState) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	var ipcReq map[string]string
	_ = json.Unmarshal([]byte(strings.TrimSpace(line)), &ipcReq)
	state.mu.Lock()
	state.requests = append(state.requests, ipcReq)
	state.mu.Unlock()
	resp := map[string]interface{}{"ok": true}
	b, _ := json.Marshal(resp)
	_, _ = conn.Write(append(b, '\n'))
	time.Sleep(10 * time.Millisecond)
}

func startGitMockIPCServer(t *testing.T, socketPath string, failFirst int) *gitIPCServerState {
	t.Helper()
	_ = os.Remove(socketPath)
	_ = os.MkdirAll(filepath.Dir(socketPath), 0755)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("mock IPC listen: %v", err)
	}
	state := &gitIPCServerState{failFirst: failFirst}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			state.mu.Lock()
			state.connectCount++
			reject := state.failFirst > 0
			if state.failFirst > 0 {
				state.failFirst--
			}
			state.mu.Unlock()
			if reject {
				conn.Close()
				continue
			}
			go handleGitIPCConn(conn, state)
		}
	}()
	t.Cleanup(func() {
		ln.Close()
		_ = os.Remove(socketPath)
	})
	return state
}

func runCLI(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	koolBin, err := resolveKoolBinary(d)
	if err != nil {
		return nil, err
	}
	args := []string{"vscode", "open-git-repo"}
	if req.RepoPath != "" {
		args = append(args, req.RepoPath)
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
	normalized, err := vscodegit.ValidateGitRepoPath(req.RepoPath, cwd)
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
	normalized, err := vscodegit.ValidateGitRepoPath(req.RepoPath, cwd)
	if err != nil {
		return &Response{ValidateErr: err.Error()}, nil
	}
	uri := vscodegit.BuildGitOpenRepoURI(normalized)
	return &Response{
		NormalizedPath: normalized,
		VSCodeURI:      uri,
	}, nil
}

func runPrecheck(t *testing.T, req *Request) (*Response, error) {
	installPrecheckHooks(t, req)
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

	installPrecheckHooks(t, req)

	socketPath := req.IPCSocketPath
	if socketPath == "" {
		socketPath = filepath.Join(t.TempDir(), "ipc.sock")
	}
	vscodegit.SetIPC_SOCKETPathForTest(socketPath)
	t.Cleanup(func() { vscodegit.SetIPC_SOCKETPathForTest("") })

	// Legacy exec phase: preserve pre-IPC behavior by forcing fallback path.
	ipcAlwaysFail := req.IPCAlwaysFail
	if req.Phase == "exec" && !req.IPCAlwaysFail && req.IPCFailConnects == 0 {
		ipcAlwaysFail = true
	}

	var server *gitIPCServerState
	if !ipcAlwaysFail {
		server = startGitMockIPCServer(t, socketPath, req.IPCFailConnects)
	}

	var stderrBuf bytes.Buffer
	vscodegit.SetStderrWriterForTest(&stderrBuf)
	t.Cleanup(func() { vscodegit.SetStderrWriterForTest(nil) })

	err := vscodegit.OpenGitRepo(req.RepoPath, cwd)
	resp := &Response{
		ExecCalled:  execCalled,
		ExecCommand: execCommand,
		ExecArgs:    execArgs,
		Stderr:      stderrBuf.String(),
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
			resp.IPCOp = server.requests[0]["op"]
			resp.IPCPath = server.requests[0]["path"]
		}
		server.mu.Unlock()
	}
	resp.StderrHint = strings.Contains(resp.Stderr, "extension not reachable via IPC")
	return resp, nil
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Phase {
	case "cli":
		return runCLI(t, d, req)
	case "validate":
		return runValidate(t, req)
	case "build-uri":
		return runBuildURI(t, req)
	case "precheck":
		return runPrecheck(t, req)
	case "ipc", "exec":
		return runOrchestrate(t, req)
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}
```