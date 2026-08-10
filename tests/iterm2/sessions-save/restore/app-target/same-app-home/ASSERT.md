## Expected

- Exit 0
- Would restore + grok resume
- Per-window **`app  ~/Applications/iTerm.app`** (create target; same-app style)
- No global prefer-home override of recorded home (create target is home)
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
	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	if !strings.Contains(out, "grok --resume "+fixtureGrokSessionID) {
		t.Fatalf("missing grok resume:\n%s", out)
	}
	if !hasSameAppCreateLineFor(out, fixtureAppHome) {
		t.Fatalf("--same-app dry-run must show app  %s create target:\n%s", fixtureAppHome, out)
	}
	// Must not show system as the create target for this window.
	if hasSameAppCreateLineFor(out, fixtureAppSystem) {
		t.Fatalf("home-recorded window must not show system create app line:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
