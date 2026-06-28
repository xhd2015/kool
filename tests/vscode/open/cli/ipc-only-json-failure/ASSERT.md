## Expected

- Exit code **1**.
- Stdout JSON has `ipc_handled: false`, normalized `path`, and non-empty `error`.
- Stderr does **not** contain URI fallback hint.
- No `vscode://` URI appears on stdout (IPC-only probe mode).

## Exit Code

- 1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 1 {
		t.Fatalf("expected exit 1, got %d\nstderr: %s\nstdout: %s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.OpenJSON == nil {
		t.Fatal("expected parsed JSON on stdout")
	}
	if resp.OpenJSON.IPCHandled {
		t.Fatal("expected ipc_handled false")
	}
	if resp.OpenJSON.Path != req.DirPath {
		t.Fatalf("JSON path=%q, want %q", resp.OpenJSON.Path, req.DirPath)
	}
	if strings.TrimSpace(resp.OpenJSON.Error) == "" {
		t.Fatalf("expected error field in JSON, got %+v", resp.OpenJSON)
	}
	if strings.Contains(resp.Stderr, "extension not reachable via IPC") {
		t.Fatalf("ipc-only must not emit URI fallback hint: %s", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "vscode://") {
		t.Fatalf("ipc-only must not invoke URI fallback, stdout: %s", resp.Stdout)
	}
}
```