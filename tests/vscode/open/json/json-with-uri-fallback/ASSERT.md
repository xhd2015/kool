## Expected

- Open succeeds (no error from `OpenDirOptions`).
- Parsed stdout JSON has `ipc_handled: false` and `fallback: "uri"` (or equivalent URI-fallback marker).
- OS opener hook **was** invoked with a `vscode://` URI.
- Stderr does **not** contain the human IPC-unreachable hint (suppressed in `--json` mode).

## Exit Code

- Success path: orchestration returns nil error.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
	if resp.ValidateErr != "" {
		t.Fatalf("expected success with URI fallback, got: %s", resp.ValidateErr)
	}
	if resp.OpenJSON == nil {
		t.Fatal("expected JSON on stdout")
	}
	if resp.OpenJSON.IPCHandled {
		t.Fatal("expected ipc_handled false when IPC unreachable")
	}
	if resp.OpenJSON.Fallback != "uri" {
		t.Fatalf("expected fallback uri in JSON, got %+v", resp.OpenJSON)
	}
	if !resp.ExecCalled {
		t.Fatal("expected OS opener after IPC failure in default mode")
	}
	if len(resp.ExecArgs) == 0 || !strings.HasPrefix(resp.ExecArgs[len(resp.ExecArgs)-1], "vscode://") {
		t.Fatalf("exec args=%v", resp.ExecArgs)
	}
	if resp.StderrHint {
		t.Fatalf("--json must suppress human fallback hint, stderr: %s", resp.Stderr)
	}
}
```