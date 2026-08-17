## Expected

- Install hook is called once with version `go1.19.13` and `req.InstallDir`.
- `ResolveGorootWith` returns the hook path.
- Prompt is written to Stderr.

## Expected Output

```
---
version: 3
---
installing go1.19.13
```

## Errors

- `err` and `resp.Err` are nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
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
	if !resp.HookCalled {
		t.Fatal("Install hook not called")
	}
	if resp.HookCount != 1 {
		t.Fatalf("Install hook count = %d, want 1", resp.HookCount)
	}
	if resp.HookVersion != "go1.19.13" {
		t.Fatalf("Install version = %q, want go1.19.13", resp.HookVersion)
	}
	if resp.HookDir != req.InstallDir {
		t.Fatalf("Install dir = %q, want %q", resp.HookDir, req.InstallDir)
	}
	if resp.Goroot != req.HookGoroot {
		t.Fatalf("ResolveGorootWith(%q) = %q, want hook path %q", req.GoVersion, resp.Goroot, req.HookGoroot)
	}
	assert.Output(t, resp.Stderr, `---
version: 3
---
installing go1.19.13
`)
}
```
