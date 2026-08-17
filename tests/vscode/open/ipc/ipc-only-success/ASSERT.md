## Expected

- `OpenDirOptions` returns **no error**.
- IPC server received an `open` request for the normalized path.
- OS opener (`exec` hook) was **not** called.
- Stderr does **not** contain URI fallback hint.

## Errors

- Any exec hook invocation means IPC-only contract was violated.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
	if resp.ValidateErr != "" {
		t.Fatalf("unexpected open error: %s", resp.ValidateErr)
	}
	if !resp.IPCCalled {
		t.Fatal("expected IPC to be called")
	}
	if resp.ExecCalled {
		t.Fatal("ipc-only success must not invoke OS opener")
	}
	if resp.StderrHint {
		t.Fatalf("unexpected stderr hint: %s", resp.Stderr)
	}
	if strings.TrimSpace(resp.IPCPath) == "" {
		t.Fatal("expected IPC path in request")
	}
}
```