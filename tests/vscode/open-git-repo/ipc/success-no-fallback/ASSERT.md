## Expected
- No orchestration error.
- IPC `git-open` op sent with normalized repo path.
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
	if resp.IPCOp != "git-open" {
		t.Fatalf("IPC op=%q, want git-open", resp.IPCOp)
	}
	if resp.IPCPath != req.RepoPath {
		t.Fatalf("IPC path=%q, want %q", resp.IPCPath, req.RepoPath)
	}
	if resp.ExecCalled {
		t.Fatal("OS opener must not be called when IPC succeeds")
	}
}
```