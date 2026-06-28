## Expected
- IPC `open` op sent with normalized absolute path.
- IPC JSON includes `"replace": true`.
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
		t.Fatal("IPC must be called")
	}
	if resp.IPCOp != "open" {
		t.Fatalf("IPC op=%q, want open", resp.IPCOp)
	}
	if resp.IPCPath != req.DirPath {
		t.Fatalf("IPC path=%q, want %q", resp.IPCPath, req.DirPath)
	}
	if !resp.IPCReplaceSet {
		t.Fatal("IPC JSON must include replace field when --replace is set")
	}
	if !resp.IPCReplace {
		t.Fatalf("IPC replace=%v, want true", resp.IPCReplace)
	}
	if resp.ExecCalled {
		t.Fatal("OS opener must not be called when IPC succeeds")
	}
}
```