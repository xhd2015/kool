## Expected

- `HandleWith` returns an error whose text contains `kool with-go`.
- Install hook is not called.

## Errors

- `resp.Err` is non-nil. Harness `err` is nil.

```go
import (
	"strings"
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
	if resp.Err == nil {
		t.Fatal("HandleWith(empty args) succeeded, want error containing kool with-go")
	}
	if !strings.Contains(resp.Err.Error(), "kool with-go") {
		t.Fatalf("HandleWith error = %q, want containing %q", resp.Err.Error(), "kool with-go")
	}
	if resp.HookCalled {
		t.Fatalf("Install hook called (version=%q dir=%q); empty args must not resolve", resp.HookVersion, resp.HookDir)
	}
}
```
