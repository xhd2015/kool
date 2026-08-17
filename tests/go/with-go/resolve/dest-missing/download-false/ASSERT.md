## Expected

- `ResolveGorootWith` returns an error (dest missing, Download false).
- Install hook is not called.
- Stderr does not receive Prompt.

## Errors

- `resp.Err` is non-nil. Harness `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err == nil {
		t.Fatalf("ResolveGorootWith(%q) succeeded with %q, want missing-dest error", req.GoVersion, resp.Goroot)
	}
	if resp.HookCalled {
		t.Fatalf("Install hook called (version=%q dir=%q); Download is false", resp.HookVersion, resp.HookDir)
	}
	if resp.Stderr != "" {
		t.Fatalf("Stderr = %q, want empty (Prompt must not run when Download is false)", resp.Stderr)
	}
}
```
