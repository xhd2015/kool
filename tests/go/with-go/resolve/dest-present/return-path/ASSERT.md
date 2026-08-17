## Expected

- `ResolveGorootWith("go1.19")` returns `$InstallDir/go1.19.13`.
- Install hook is not called.
- Stderr does not receive Prompt (dest already exists).

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
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("ResolveGorootWith(%q) failed: %v", req.GoVersion, resp.Err)
	}
	want := destPin(req.InstallDir)
	if resp.Goroot != want {
		t.Fatalf("ResolveGorootWith(%q) = %q, want %q", req.GoVersion, resp.Goroot, want)
	}
	if resp.HookCalled {
		t.Fatalf("Install hook called (version=%q dir=%q); dest already existed", resp.HookVersion, resp.HookDir)
	}
	if resp.Stderr != "" {
		t.Fatalf("Stderr = %q, want empty (Prompt must not run when dest exists)", resp.Stderr)
	}
}
```
