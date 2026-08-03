## Expected

- Exit 0
- stderr already-running warn for grok + pid
- restored_at stamped
- AS called; scripts contain mark resume; **no** grok --resume for skipped id
- stdout Restored summary mentions skip count when skipped>0

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
	if !strings.Contains(errOut, "warning:") ||
		!strings.Contains(strings.ToLower(errOut), "already running") {
		t.Fatalf("expected already-running warn; stderr=%q", errOut)
	}
	if !strings.Contains(errOut, fixtureGrokSessionID) {
		t.Fatalf("warn must name grok session; stderr=%q", errOut)
	}
	if !strings.Contains(strings.ToLower(errOut), "pid") {
		t.Fatalf("warn must include pid; stderr=%q", errOut)
	}

	if resp.Doc == nil || !resp.Doc.IsConsumed() {
		t.Fatal("live partial restore must stamp restored_at")
	}

	if resp.RestoreASCallCount < 1 {
		t.Fatalf("expected AS for remaining tabs; callCount=%d scripts=%v",
			resp.RestoreASCallCount, resp.RestoreASScripts)
	}
	joined := strings.Join(resp.RestoreASScripts, "\n")
	if !strings.Contains(joined, "create window") {
		t.Fatalf("AS should create window for remaining tabs:\n%s", joined)
	}
	// Remaining mark must be resumed.
	if !strings.Contains(joined, "mark") {
		t.Fatalf("AS must include mark resume for non-skipped tab:\n%s", joined)
	}
	// Skipped grok must not be in AS resume commands.
	if strings.Contains(joined, "grok --resume "+fixtureGrokSessionID) {
		t.Fatalf("skipped grok must not appear in restore AS:\n%s", joined)
	}

	out := strings.ToLower(resp.Stdout)
	if !strings.Contains(out, "restored") {
		t.Fatalf("expected Restored summary:\n%s", resp.Stdout)
	}
	// E4: skip clause when skipped > 0
	if !strings.Contains(out, "skip") {
		t.Fatalf("summary must mention skip count when skipped>0:\n%s", resp.Stdout)
	}
}
```
