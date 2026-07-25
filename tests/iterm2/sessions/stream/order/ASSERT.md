## Expected

- Exit 0.
- `LastListTabsInvoked` true (fixture has two windows).
- **`SawW1BeforeLastListTabs` true** — progressive stream wrote `W1` before
  collecting the last window.
- Final stdout contains `W1`, `W2`, and a sessions summary footer
  (e.g. `2 sessions` and idle/busy counts).

## Errors

- None.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.LastListTabsInvoked {
		t.Fatal("expected ListTabsAndSessions for last window (probe not fired; Capture not phased?)")
	}
	if !resp.SawW1BeforeLastListTabs {
		t.Fatalf("expected W1 on stdout before last ListTabs (streaming); stdout now:\n%s", resp.Stdout)
	}
	out := resp.Stdout
	for _, want := range []string{"W1", "W2", "sessions"} {
		if !strings.Contains(out, want) {
			t.Fatalf("final stdout missing %q:\n%s", want, out)
		}
	}
}
```
