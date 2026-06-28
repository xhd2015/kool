## Expected
- No validation or orchestration error.
- IPC `open` op sent with normalized absolute path.
- IPC JSON omits `replace` field (default new-window behavior).
- OS opener exec hook NOT called.

## Exit Code
- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ValidateErr != "" {
		t.Fatalf("unexpected open error: %s", resp.ValidateErr)
	}
	if !resp.IPCCalled {
		t.Fatal("IPC must be called on success path")
	}
	if resp.IPCOp != "open" {
		t.Fatalf("IPC op=%q, want open", resp.IPCOp)
	}
	if resp.IPCPath != req.DirPath {
		t.Fatalf("IPC path=%q, want %q", resp.IPCPath, req.DirPath)
	}
	if resp.IPCReplaceSet {
		t.Fatal("IPC replace field must be omitted for default open")
	}
	if resp.ExecCalled {
		t.Fatal("OS opener must not be called when IPC succeeds")
	}
}
```