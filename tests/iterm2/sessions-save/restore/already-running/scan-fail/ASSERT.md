## Expected

- Exit 0 (soft fail — not hard Error abort)
- stderr: soft `warning:` about scan/capture/snapshot failure (not only host)
- no `already running` skip warnings
- stdout: full Would restore plan with both resume cmds
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
		t.Fatalf("scan fail must be soft (exit 0); exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	errOut := resp.Stderr
	if !strings.Contains(errOut, "warning:") {
		t.Fatalf("capture fail must soft-warn on stderr; stderr=%q", errOut)
	}
	// Must not hard-fail the restore with Error: as the primary path.
	// Require a scan/capture-related soft warn (wording flexible).
	lowErr := strings.ToLower(errOut)
	scanish := strings.Contains(lowErr, "scan") ||
		strings.Contains(lowErr, "capture") ||
		strings.Contains(lowErr, "snapshot") ||
		strings.Contains(lowErr, "iterm") ||
		strings.Contains(lowErr, "live session") ||
		(strings.Contains(lowErr, "not running") || strings.Contains(lowErr, "not available"))
	if !scanish {
		t.Fatalf("expected soft warn about failed live scan/capture; stderr=%q", errOut)
	}
	if strings.Contains(lowErr, "already running") {
		t.Fatalf("scan fail must not produce already-running skip hits; stderr=%q", errOut)
	}

	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing full Would restore plan:\n%s", out)
	}
	if !strings.Contains(out, "grok --resume "+fixtureGrokSessionID) {
		t.Fatalf("scan fail restores all — missing grok:\n%s", out)
	}
	if !strings.Contains(out, "mark '"+fixtureMarkMessage+"'") &&
		!strings.Contains(out, fixtureMarkMessage) {
		t.Fatalf("scan fail restores all — missing mark:\n%s", out)
	}
	// Full counts (0 hits).
	if !strings.Contains(out, "2 tabs") {
		t.Fatalf("scan fail should would-restore all 2 tabs:\n%s", out)
	}
	lowOut := strings.ToLower(out)
	if strings.Contains(lowOut, "skip (already running)") ||
		strings.Contains(lowOut, "skipped already running") {
		t.Fatalf("scan fail must not skip tabs as already running:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
