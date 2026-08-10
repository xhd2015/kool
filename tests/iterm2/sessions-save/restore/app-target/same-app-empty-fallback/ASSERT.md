## Expected

- Exit 0 (warn, not hard error)
- stderr contains `warning:` (empty/missing app under `--same-app`)
- Would restore continues
- Fallback create target is home when both installs exist (prefer-home):
  either global `restore target` home, or same-app `app  ~/Applications/...`
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
	if !strings.Contains(errOut, "warning") {
		t.Fatalf("--same-app empty app must warn; stderr=%q", resp.Stderr)
	}
	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	// Prefer-home fallback when both installs exist (R7).
	homeOK := hasRestoreTargetLine(out, fixtureAppHome) ||
		hasSameAppCreateLineFor(out, fixtureAppHome) ||
		strings.Contains(out, fixtureAppHome)
	if !homeOK {
		t.Fatalf("empty app under --same-app should fall back to home when both exist:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
