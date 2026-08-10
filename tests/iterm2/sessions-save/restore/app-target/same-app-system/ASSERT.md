## Expected

- Exit 0
- restored_at stamped
- MockRestoreAS called ≥1
- AS path-tells system install (`/Applications/iTerm.app`), not only bare `"iTerm2"`
- Stdout Restored summary (or equivalent success)

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
	if resp.Doc == nil || !resp.Doc.IsConsumed() {
		t.Fatal("live same-app restore must stamp restored_at")
	}
	if resp.RestoreASCallCount < 1 {
		t.Fatalf("expected AS scripts; callCount=%d", resp.RestoreASCallCount)
	}
	joined := strings.Join(resp.RestoreASScripts, "\n")
	// Path tell to system install (R1 / path-tell contract).
	if !strings.Contains(joined, fixtureAppSystem) &&
		!strings.Contains(joined, `application "/Applications/iTerm.app"`) {
		t.Fatalf("--same-app system must path-tell %s in AS (not only bare iTerm2):\n%s",
			fixtureAppSystem, joined)
	}
	// Must not be exclusively bare tell when system path is the create target.
	// Bare may appear as last-resort fallback only; with recorded system app it must path-tell.
	if strings.Contains(joined, `tell application "iTerm2"`) &&
		!strings.Contains(joined, "Applications/iTerm.app") {
		t.Fatalf("expected path tell, got bare iTerm2 only:\n%s", joined)
	}
	out := strings.ToLower(resp.Stdout)
	if !strings.Contains(out, "restored") {
		t.Fatalf("expected Restored summary:\n%s", resp.Stdout)
	}
}
```
