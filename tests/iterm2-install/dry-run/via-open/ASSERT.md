## Expected

- Exit 0.
- Stdout contains dry-run banner and resolved version `3.6.11`.
- Stdout plan reflects via-open mode: at least one of `via-open`, `user open`,
  `user-open`, `open`, `clear-quarantine`, `quarantine` in a mode/steps context
  (case-insensitive search on full stdout is acceptable for P1).
- No zip written under `--download-dir`.

## Side Effects

- No zip file at planned path under DownloadDir.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `<contains>
dry-run
3.6.11
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	mentionsOpenMode := strings.Contains(low, "via-open") ||
		strings.Contains(low, "user open") ||
		strings.Contains(low, "user-open") ||
		strings.Contains(low, "clear-quarantine") ||
		strings.Contains(low, "quarantine") ||
		// bare "open" as a step token (avoid matching only unrelated words if possible)
		strings.Contains(low, " open") ||
		strings.Contains(low, "open,") ||
		strings.Contains(low, "steps:") && strings.Contains(low, "open")
	if !mentionsOpenMode {
		t.Fatalf("dry-run --via-open plan should mention open/via-open/quarantine; got:\n%s", resp.Stdout)
	}
	assertNoZip(t, resp)
}
```
