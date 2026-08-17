## Expected

- `HandleWith` succeeds (`/usr/bin/true`).
- Install hook is not called (GOROOT= skips resolve).
- Stderr does not receive Prompt.

## Errors

- `err` and `resp.Err` are nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("HandleWith(GOROOT=) failed: %v", resp.Err)
	}
	if resp.HookCount != 0 {
		t.Fatalf("Install hook count = %d, want 0 (GOROOT= skips resolve)", resp.HookCount)
	}
	if resp.HookCalled {
		t.Fatalf("Install hook called (version=%q dir=%q); GOROOT= must skip resolve", resp.HookVersion, resp.HookDir)
	}
	if resp.Stderr != "" {
		t.Fatalf("Stderr = %q, want empty (Prompt must not run when GOROOT= skips resolve)", resp.Stderr)
	}
}
```
