## Expected

- Exit 0
- stderr: `warning:` + `already running` + mark message + `pid`
- stdout: reduced remaining counts (1 tab); skip mark; grok resume still planned
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
	if !strings.Contains(errOut, fixtureMarkMessage) {
		t.Fatalf("warning must identify mark message; stderr=%q", errOut)
	}
	if !strings.Contains(strings.ToLower(errOut), "pid") {
		t.Fatalf("warning must include pid; stderr=%q", errOut)
	}

	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	if strings.Contains(out, "2 tabs") {
		t.Fatalf("header must not count skipped tab as would-create:\n%s", out)
	}
	low := strings.ToLower(out)
	if !strings.Contains(low, "skip") || !strings.Contains(low, "already running") {
		t.Fatalf("plan must mark skip (already running):\n%s", out)
	}
	if !strings.Contains(out, "grok --resume "+fixtureGrokSessionID) {
		t.Fatalf("remaining grok would restore missing resume:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
