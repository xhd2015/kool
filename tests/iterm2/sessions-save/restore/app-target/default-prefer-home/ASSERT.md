## Expected

- Exit 0
- Would restore + grok resume
- Global **`restore target`** is home (`~/Applications/iTerm.app`)
- Per-window **`recorded app`** is system (`/Applications/iTerm.app`) because it differs
- Default mode: no same-app create line `app  …` (distinct from `recorded app`)
- restored_at still null

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
	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	if !strings.Contains(out, "grok --resume "+fixtureGrokSessionID) {
		t.Fatalf("missing grok resume:\n%s", out)
	}
	// Global restore target prefers home when both installs exist (R2).
	if !hasRestoreTargetLine(out, fixtureAppHome) {
		t.Fatalf("expected restore target %s when both installs on disk:\n%s", fixtureAppHome, out)
	}
	// Must not claim system as the global restore target.
	if hasRestoreTargetLine(out, fixtureAppSystem) {
		t.Fatalf("default with both installs must not target system:\n%s", out)
	}
	// Recorded app shown because it differs from restore target (Open2).
	if !hasRecordedAppLine(out, fixtureAppSystem) {
		t.Fatalf("expected recorded app %s when it differs from target:\n%s", fixtureAppSystem, out)
	}
	// Default does not use per-window create `app  ` line (that is --same-app).
	if hasSameAppCreateLine(out) {
		t.Fatalf("default mode must not print same-app create `app  ` lines:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```