# Scenario

**Feature**: kool opens directories in VS Code via IPC with URI fallback

```
# CLI validates dir, prechecks VS Code, tries IPC, falls back to URI
kool vscode open <dir> -> ValidateDirPath -> EnsureCodeCLI -> EnsureExtensionListed

# IPC-first open; OS opener only on failure
EnsureExtensionListed -> IPC client (open op) -> VS Code extension
IPC client (fail) -> OS opener (vscode://.../open?path=...) -> VS Code
```

## Preconditions
- Go toolchain available.
- `github.com/xhd2015/kool/vscodegit` package provides testable functions and hooks (implementer adds).
- CLI integration tests require built `kool` binary in PATH or module root.

## Steps
1. Each leaf sets `req.Phase` and path fields via `Setup`.
2. `Run` dispatches to validate, URI build, precheck, IPC orchestration, or CLI subprocess.
3. Each leaf asserts outcomes via `Assert`.

## Context
- **Extension id**: `xhd2015.open-in-new-window`
- **URI path**: `/open?path=<url-encoded-absolute-path>`; append `&replace=true` when `--replace`
- **IPC socket**: `~/.kool/xhd2015.open-in-new-window.sock` (overridable in tests)
- **IPC fallback stderr**: `Note: extension not reachable via IPC; opening via vscode:// URI.`
- **CLI mock IPC**: subprocess tests set `KOOL_VSCODE_IPC_SOCKET` to a mock Unix socket
  started in the leaf `Setup` (implementer reads this env in `ipcSocketPath()`).
- **Relative paths**: resolved against cwd before validation

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func expectedOpenURI(absPath string, replace bool) string {
	encoded := strings.ReplaceAll(absPath, " ", "%20")
	uri := fmt.Sprintf("vscode://xhd2015.open-in-new-window/open?path=%s", encoded)
	if replace {
		uri += "&replace=true"
	}
	return uri
}

func Setup(t *testing.T, req *Request) error {
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	if err := os.Chdir(req.WorkingDir); err != nil {
		return err
	}
	return nil
}

func osMkdir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func writeFakeCodeScript(t *testing.T, dir string, listExtensions []string) string {
	t.Helper()
	script := filepath.Join(dir, "code")
	body := `#!/bin/sh
case "$1" in
--list-extensions)
`
	for _, ext := range listExtensions {
		body += fmt.Sprintf("  echo '%s'\n", ext)
	}
	body += `  ;;
*)
  echo "unknown: $1" >&2
  exit 1
  ;;
esac
`
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatalf("write fake code: %v", err)
	}
	return script
}

func writeFakeCodeMissing(t *testing.T, dir string) string {
	t.Helper()
	script := filepath.Join(dir, "code")
	body := `#!/bin/sh
echo "code: command not found" >&2
exit 127
`
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatalf("write fake code: %v", err)
	}
	return script
}

func initValidDir(t *testing.T, baseDir string, name string) string {
	t.Helper()
	dir := filepath.Join(baseDir, name)
	if err := osMkdir(dir); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	marker := filepath.Join(dir, ".keep")
	if err := os.WriteFile(marker, []byte("ok"), 0644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	return dir
}

func installExtensionListedPrecheck(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.WorkingDir, "bin")
	if err := osMkdir(binDir); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	req.CodeCommand = writeFakeCodeScript(t, binDir, []string{"xhd2015.open-in-new-window"})
	req.CodeInPath = true
}

func installCLIMockIPC(t *testing.T, req *Request) {
	t.Helper()
	socketPath := filepath.Join(req.WorkingDir, "ipc-cli.sock")
	if req.IPCSocketPath != "" {
		socketPath = req.IPCSocketPath
	}
	req.IPCSocketPath = socketPath
	startMockIPCServer(t, socketPath, req.IPCFailConnects)
}

func installNoExtensionPrecheck(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.WorkingDir, "bin")
	if err := osMkdir(binDir); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	req.CodeCommand = writeFakeCodeScript(t, binDir, []string{"other.extension"})
	req.CodeInPath = true
}

// markRootTree keeps hierarchical child packages importing this package live.
func markRootTree() {}
```