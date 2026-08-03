## Expected

- Exit 0
- stderr already-running warn for grok
- header remaining: 1 window / 1 tab (not 2 tabs)
- body: skip line for hit + would-restore/resume for mark
- skip meta when skipped>0
- not stamped

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
	errOut := strings.ToLower(resp.Stderr)
	if !strings.Contains(resp.Stderr, "warning:") || !strings.Contains(errOut, "already running") {
		t.Fatalf("expected already-running warn; stderr=%q", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, fixtureGrokSessionID) {
		t.Fatalf("warn must name grok session; stderr=%q", resp.Stderr)
	}

	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	if strings.Contains(out, "2 tabs") {
		t.Fatalf("mixed window header must use remaining would-create tab count:\n%s", out)
	}
	if !strings.Contains(out, "1 tabs") && !strings.Contains(out, "1 tab") {
		t.Fatalf("expected 1 remaining tab in header:\n%s", out)
	}
	if !strings.Contains(out, "1 windows") && !strings.Contains(out, "1 window") {
		t.Fatalf("window with remaining tabs still planned (1 window):\n%s", out)
	}
	low := strings.ToLower(out)
	if !strings.Contains(low, "skip") || !strings.Contains(low, "already running") {
		t.Fatalf("body/meta must include skip (already running):\n%s", out)
	}
	if !strings.Contains(out, "mark '"+fixtureMarkMessage+"'") &&
		!strings.Contains(out, fixtureMarkMessage) {
		t.Fatalf("miss tab must still show would-restore mark resume:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
