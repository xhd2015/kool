## Expected

- Exit 0
- stderr already-running warnings (both identities)
- restored_at stamped even when 0 tabs restored (E1)
- no `create window` in AS scripts (AS not called or empty of creates)
- summary: 0 restored + skipped count

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
		t.Fatalf("expected already-running warnings; stderr=%q", errOut)
	}
	if !strings.Contains(errOut, fixtureGrokSessionID) {
		t.Fatalf("warn must mention grok; stderr=%q", errOut)
	}
	if !strings.Contains(errOut, fixtureMarkMessage) {
		t.Fatalf("warn must mention mark message; stderr=%q", errOut)
	}

	if resp.Doc == nil || !resp.Doc.IsConsumed() {
		t.Fatal("live all-skip must still stamp restored_at (E1)")
	}

	joined := strings.Join(resp.RestoreASScripts, "\n")
	if strings.Contains(joined, "create window") {
		t.Fatalf("all-skip must not create windows via AS:\n%s", joined)
	}
	// Prefer zero AS calls; tolerate empty scripts without create.
	if resp.RestoreASCallCount > 0 {
		for i, sc := range resp.RestoreASScripts {
			if strings.Contains(sc, "create window") || strings.Contains(sc, "write text") {
				t.Fatalf("AS[%d] must not restore tabs when all skipped:\n%s", i, sc)
			}
		}
	}

	out := strings.ToLower(resp.Stdout)
	if !strings.Contains(out, "restored") {
		t.Fatalf("expected Restored summary:\n%s", resp.Stdout)
	}
	// 0 restored + skip clause
	if !strings.Contains(out, "0") {
		t.Fatalf("summary should reflect 0 restored tabs/windows:\n%s", resp.Stdout)
	}
	if !strings.Contains(out, "skip") {
		t.Fatalf("summary must mention skipped when all skipped:\n%s", resp.Stdout)
	}
}
```
