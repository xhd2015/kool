## Expected

- `ResolveGorootWith("1.19")` returns a path whose last element is `go1.19.13`.
- Install hook is unused (dest-exists fixture).

## Errors

- `err` and `resp.Err` are nil.

```go
import (
	"path/filepath"
	"strings"
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
	if !strings.HasSuffix(resp.Goroot, "go1.19.13") {
		t.Fatalf("ResolveGorootWith(%q) = %q, want suffix go1.19.13", req.GoVersion, resp.Goroot)
	}
	if filepath.Base(resp.Goroot) != "go1.19.13" {
		t.Fatalf("ResolveGorootWith(%q) base = %q, want go1.19.13", req.GoVersion, filepath.Base(resp.Goroot))
	}
	if resp.HookCalled {
		t.Fatalf("Install hook called (version=%q dir=%q); dest already existed", resp.HookVersion, resp.HookDir)
	}
}
```
