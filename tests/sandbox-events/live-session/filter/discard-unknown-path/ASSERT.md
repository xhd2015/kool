## Expected

- Build ok; sock existed (session was live and could receive/discard).
- Notify ran (publisher may still exit 0 even when sessions discard).
- Session `reload-me.txt` still contains `old-filter-content`.
- Session file must **not** contain `new-filter-content` (would mean wrongly
  reloaded the rebuilt load despite notify path mismatch).

## Exit Code

- build: 0

```go
import (
	"strings"
	"testing"

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
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.SockExistsAfterStart {
		// Without a live listener the filter behavior cannot be proven; RED
		// until sock bind exists (same gate as hot-reload).
		t.Fatalf("filter leaf needs live events sock; parent=%q", resp.SandboxRootParent)
	}
	if !resp.NotifyRan {
		t.Fatal("expected notify-event for unknown path")
	}
	if resp.SessionRoot == "" || !resp.SessionFileExists {
		t.Fatalf("expected session file snapshot; root=%q exists=%v content=%q",
			resp.SessionRoot, resp.SessionFileExists, resp.SessionFileContent)
	}
	if !strings.Contains(resp.SessionFileContent, "old-filter-content") {
		t.Fatalf("session file should remain old-filter-content; got %q", resp.SessionFileContent)
	}
	if strings.Contains(resp.SessionFileContent, "new-filter-content") {
		t.Fatalf("session must not apply load for unknown notify path; content=%q",
			resp.SessionFileContent)
	}
}
```
