## Expected

- Exit 0.
- StartSession called with Teardown=true.
- Stop called after WaitSignal.

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.StartCalled {
		t.Fatal("expected StartSession")
	}
	if !resp.StartTeardown {
		t.Fatal("expected StartSession Teardown=true by default")
	}
	if !resp.StopCalled {
		t.Fatal("expected Session.Stop")
	}
}
```
