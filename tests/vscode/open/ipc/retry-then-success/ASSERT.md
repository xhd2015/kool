## Expected
- More than one IPC connect attempt.
- IPC eventually succeeds with `open` op.
- OS opener NOT called.

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
	if resp.IPCAttempts < 2 {
		t.Fatalf("expected at least 2 IPC connect attempts, got %d", resp.IPCAttempts)
	}
	if !resp.IPCCalled {
		t.Fatal("IPC open must succeed after retry")
	}
	if resp.IPCOp != "open" {
		t.Fatalf("IPC op=%q, want open", resp.IPCOp)
	}
	if resp.ExecCalled {
		t.Fatal("OS opener must not be called when IPC eventually succeeds")
	}
}
```