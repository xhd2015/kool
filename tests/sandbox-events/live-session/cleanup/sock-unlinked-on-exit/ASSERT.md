## Expected

- Build ok.
- Sock existed after start (`SockExistsAfterStart`).
- Guest exited (`GuestExited`).
- Sock does not exist after exit (`SockExistsAfterExit` false).

## Exit Code

- build: 0

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
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.SockExistsAfterStart {
		// RED until listener binds; cleanup cannot be proven without bind.
		t.Fatalf("expected sock after start before cleanup check; parent=%q",
			resp.SandboxRootParent)
	}
	if !resp.GuestExited {
		t.Fatal("expected guest to exit for cleanup assert")
	}
	if resp.SockExistsAfterExit {
		t.Fatalf("sock should be unlinked after session end; path=%q", resp.SockPath)
	}
}
```
