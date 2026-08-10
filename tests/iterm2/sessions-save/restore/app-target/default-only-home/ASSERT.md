## Expected

- Exit 0
- Would restore
- **`restore target`** is home (`~/Applications/iTerm.app`)
- Does **not** claim system as restore target
- **`recorded app`** system when it differs
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
	if !hasRestoreTargetLine(out, fixtureAppHome) {
		t.Fatalf("expected restore target %s when only home on disk:\n%s", fixtureAppHome, out)
	}
	if hasRestoreTargetLine(out, fixtureAppSystem) {
		t.Fatalf("only-home disk must not claim system restore target:\n%s", out)
	}
	if !hasRecordedAppLine(out, fixtureAppSystem) {
		t.Fatalf("expected recorded app %s when it differs from home target:\n%s", fixtureAppSystem, out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
