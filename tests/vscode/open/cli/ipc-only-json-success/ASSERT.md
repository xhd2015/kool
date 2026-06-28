## Expected Output

```json
{"ipc_handled":true,"path":"<normalized-dir>"}
```

## Expected

- Exit code **0**.
- Stdout is a single JSON object with `ipc_handled: true` and `path` equal to the normalized directory path.
- Stderr does **not** contain the URI fallback hint (`extension not reachable via IPC`).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d\nstderr: %s\nstdout: %s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.OpenJSON == nil {
		t.Fatal("expected parsed JSON on stdout")
	}
	if !resp.OpenJSON.IPCHandled {
		t.Fatalf("expected ipc_handled true, got %+v", resp.OpenJSON)
	}
	if resp.OpenJSON.Path != req.DirPath {
		t.Fatalf("JSON path=%q, want %q", resp.OpenJSON.Path, req.DirPath)
	}
	if strings.Contains(resp.Stderr, "extension not reachable via IPC") {
		t.Fatalf("unexpected URI fallback hint on stderr: %s", resp.Stderr)
	}
}
```