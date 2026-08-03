## Expected

- Exit 0
- stderr: `warning:` + `already running` + grok session_id + `pid`
- stdout: Would restore with **reduced** remaining tab count (1, not 2)
- stdout: skip marker for already-running tab; would-restore path for mark resume
- skip meta when skipped>0 (`skipped` + `already running` or equivalent)
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
	errOut := resp.Stderr
	if !strings.Contains(errOut, "warning:") {
		t.Fatalf("expected already-running warning; stderr=%q", errOut)
	}
	if !strings.Contains(strings.ToLower(errOut), "already running") {
		t.Fatalf("warning must say already running; stderr=%q", errOut)
	}
	if !strings.Contains(errOut, fixtureGrokSessionID) {
		t.Fatalf("warning must identify grok session_id; stderr=%q", errOut)
	}
	if !strings.Contains(strings.ToLower(errOut), "pid") {
		t.Fatalf("warning must include pid; stderr=%q", errOut)
	}

	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	// Header counts = would-create only (1 tab remaining, not 2).
	if strings.Contains(out, "2 tabs") {
		t.Fatalf("header must not count skipped tab as would-create (got 2 tabs):\n%s", out)
	}
	if !strings.Contains(out, "1 tabs") && !strings.Contains(out, "1 tab") {
		t.Fatalf("header should show 1 remaining tab to create:\n%s", out)
	}
	low := strings.ToLower(out)
	if !strings.Contains(low, "skip") || !strings.Contains(low, "already running") {
		t.Fatalf("plan body/meta must mark skip (already running):\n%s", out)
	}
	// Remaining mark tab still planned.
	if !strings.Contains(out, "mark '"+fixtureMarkMessage+"'") &&
		!strings.Contains(out, "mark "+fixtureMarkMessage) {
		t.Fatalf("remaining mark would restore missing resume:\n%s", out)
	}
	// Skipped grok must not appear as a would-restore resume line alone without skip.
	// Soft: require skip meta; do not require absence of session id in skip line.
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
