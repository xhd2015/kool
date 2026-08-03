## Expected

- Exit 0
- stderr: no `already running` warning (host mismatch soft-ok)
- stdout: Would restore full counts (2 tabs); both resume cmds
- no skip meta for already-running
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
	if strings.Contains(errOut, "already running") {
		t.Fatalf("unrelated live panes must not warn already running; stderr=%q", resp.Stderr)
	}

	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	// Full would-create counts (scan ran, 0 hits).
	if !strings.Contains(out, "2 tabs") {
		t.Fatalf("miss must would-restore all 2 tabs:\n%s", out)
	}
	if !strings.Contains(out, "grok --resume "+fixtureGrokSessionID) {
		t.Fatalf("missing grok resume:\n%s", out)
	}
	if !strings.Contains(out, "mark '"+fixtureMarkMessage+"'") &&
		!strings.Contains(out, fixtureMarkMessage) {
		t.Fatalf("missing mark resume:\n%s", out)
	}
	low := strings.ToLower(out)
	// Must not claim skips when none.
	if strings.Contains(low, "skipped already running") ||
		strings.Contains(low, "skip (already running)") {
		t.Fatalf("miss must not mark any tab skip (already running):\n%s", out)
	}
	// After feature: tab-level would-restore action markers in addition to the
	// header "Would restore" (force RED until implementer adds per-tab actions).
	if strings.Count(low, "would restore") < 2 {
		t.Fatalf("expected tab-level would-restore action markers (count>=2 incl header):\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
