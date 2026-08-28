## Expected

- Exit 0.
- StartSession + WaitSignal + Stop called.
- WaitReady not called.
- Stdout has Ctrl+C and does not mention Checking public readiness.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.StartCalled {
		t.Fatal("expected StartSession")
	}
	if resp.WaitReadyCalled {
		t.Fatal("WaitReady must not run with --no-wait-ready")
	}
	if !resp.WaitSignalCalled {
		t.Fatal("expected WaitSignal")
	}
	if !resp.StopCalled {
		t.Fatal("expected Stop")
	}
	if strings.Contains(resp.Stdout, "Checking public readiness") {
		t.Fatalf("stdout must not check readiness: %q", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `<contains>
https://a.example.com
Press Ctrl+C to stop
</contains>
`)
}
```
