## Expected

- Exit 0
- stderr: already-running warn identifies **codex** session_id + pid
- stdout: 0 would-create tabs; skip marker; not stamped
- does not require grok wording for this tab

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
	if !strings.Contains(errOut, fixtureCodexSessionID) {
		t.Fatalf("warning must identify codex session_id; stderr=%q", errOut)
	}
	// Kind-aware key: mention codex (not only generic session).
	if !strings.Contains(strings.ToLower(errOut), "codex") {
		t.Fatalf("warning should identify kind=codex; stderr=%q", errOut)
	}
	if !strings.Contains(strings.ToLower(errOut), "pid") {
		t.Fatalf("warning must include pid; stderr=%q", errOut)
	}

	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	if !strings.Contains(out, "0 tabs") && !strings.Contains(out, "0 tab") {
		t.Fatalf("all-hit codex dry-run must show 0 would-create tabs:\n%s", out)
	}
	low := strings.ToLower(out)
	if !strings.Contains(low, "skip") || !strings.Contains(low, "already running") {
		t.Fatalf("plan must mark skip (already running):\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
