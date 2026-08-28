## Expected

- Exit 0.
- WaitReady called; WaitSignal and Stop still called.
- Stderr contains `warning: public ready timeout`.
- Stdout still reaches `Press Ctrl+C to stop`.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
	"time"

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
	if !resp.WaitReadyCalled {
		t.Fatal("expected WaitReady")
	}
	if resp.WaitReadyTimeout != 5*time.Second {
		t.Fatalf("WaitReadyTimeout=%v want 5s", resp.WaitReadyTimeout)
	}
	if !resp.WaitSignalCalled {
		t.Fatal("expected WaitSignal after timeout warning")
	}
	if !resp.StopCalled {
		t.Fatal("expected Stop after WaitSignal")
	}
	if !strings.Contains(resp.Stderr, "warning: public ready timeout") {
		t.Fatalf("stderr missing timeout warning: %q", resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `<contains>
Press Ctrl+C to stop
</contains>
`)
}
```
