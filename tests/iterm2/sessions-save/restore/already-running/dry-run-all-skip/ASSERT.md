## Expected

- Exit 0
- stderr: already-running warnings for both tabs (or at least grok + mark ids)
- stdout: Would restore with **0** would-create windows/tabs
- skip lines / skip meta; no create resume as would-restore without skip
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
		t.Fatalf("expected already-running warnings; stderr=%q", errOut)
	}
	if !strings.Contains(strings.ToLower(errOut), "already running") {
		t.Fatalf("stderr must say already running; stderr=%q", errOut)
	}
	if !strings.Contains(errOut, fixtureGrokSessionID) {
		t.Fatalf("warn must mention grok session; stderr=%q", errOut)
	}
	if !strings.Contains(errOut, fixtureMarkMessage) {
		t.Fatalf("warn must mention mark message; stderr=%q", errOut)
	}

	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore plan header:\n%s", out)
	}
	// 0 would-create.
	if !strings.Contains(out, "0 windows") && !strings.Contains(out, "0 window") {
		t.Fatalf("all-skip header must show 0 windows would create:\n%s", out)
	}
	if !strings.Contains(out, "0 tabs") && !strings.Contains(out, "0 tab") {
		t.Fatalf("all-skip header must show 0 tabs would create:\n%s", out)
	}
	low := strings.ToLower(out)
	if !strings.Contains(low, "skip") || !strings.Contains(low, "already running") {
		t.Fatalf("plan must list skip (already running):\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
