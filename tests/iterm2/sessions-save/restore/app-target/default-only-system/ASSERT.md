## Expected

- Exit 0
- Would restore
- **`restore target`** is system (`/Applications/iTerm.app`) — only install on disk
- Does **not** claim home as restore target
- **`recorded app`** home may appear (differs from target)
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
	if !hasRestoreTargetLine(out, fixtureAppSystem) {
		t.Fatalf("expected restore target %s when only system on disk:\n%s", fixtureAppSystem, out)
	}
	if hasRestoreTargetLine(out, fixtureAppHome) {
		t.Fatalf("only-system disk must not claim home restore target:\n%s", out)
	}
	// Recorded home differs → may show recorded app line (Open2).
	if !hasRecordedAppLine(out, fixtureAppHome) {
		t.Fatalf("expected recorded app %s when it differs from system target:\n%s", fixtureAppHome, out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
