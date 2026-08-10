## Expected

- Exit 0
- `LastListTabsInvoked` true (fixture has two windows)
- **`SawW1BeforeLastListTabs` true** — progressive stream wrote `W1` before
  collecting the last window
- Final stdout contains `W1`, `W2`, and Would save footer
- No file written

## Errors

- None

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
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
	if strings.HasPrefix(out, "\n") {
		t.Fatalf("stream dry-run must not start with a leading empty line; stdout:\n%q", out)
	}
	for _, want := range []string{"W1", "W2", "Would save"} {
		if !strings.Contains(out, want) {
			t.Fatalf("final stdout missing %q:\n%s", want, out)
		}
	}
	// Separator blank only between windows (after W1 block, before W2).
	if !strings.Contains(out, "\n\n  ") {
		t.Fatalf("expected blank line between W1 and W2; stdout:\n%q", out)
	}
	p := filepath.Join(req.WorkingDir, "stream-plan.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("dry-run must not write %s", p)
	}
}
```
