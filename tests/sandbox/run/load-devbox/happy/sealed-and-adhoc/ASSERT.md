## Expected

- Build + secondaries ok; sealed run exit 0.
- Two notices: sealed A then adhoc B (path list order: sealed pack first, then CLI).
- Guest stdout includes both `content-sealed-a` and `content-adhoc-b`.

## Exit Code

- sealed run: 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if len(resp.SecondaryPaths) != 2 {
		t.Fatalf("expected 2 secondaries; got %v", resp.SecondaryPaths)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	absA := resp.SecondaryPaths[0]
	absB := resp.SecondaryPaths[1]
	noticeA := "notice: loading devbox " + absA
	noticeB := "notice: loading devbox " + absB
	if !strings.Contains(resp.RunStdout, noticeA) {
		t.Fatalf("want notice for sealed A %q; stdout=%q", noticeA, resp.RunStdout)
	}
	if !strings.Contains(resp.RunStdout, noticeB) {
		t.Fatalf("want notice for adhoc B %q; stdout=%q", noticeB, resp.RunStdout)
	}
	iA := strings.Index(resp.RunStdout, noticeA)
	iB := strings.Index(resp.RunStdout, noticeB)
	if iA < 0 || iB < 0 || iA > iB {
		t.Fatalf("notices should appear sealed-then-adhoc; iA=%d iB=%d stdout=%q", iA, iB, resp.RunStdout)
	}
	// Exactly one notice line per path (no duplicate for each).
	if strings.Count(resp.RunStdout, noticeA) != 1 {
		t.Fatalf("want exactly one notice for A; stdout=%q", resp.RunStdout)
	}
	if strings.Count(resp.RunStdout, noticeB) != 1 {
		t.Fatalf("want exactly one notice for B; stdout=%q", resp.RunStdout)
	}
	if !strings.Contains(resp.RunStdout, "content-sealed-a") {
		t.Fatalf("missing sealed-a content; stdout=%q", resp.RunStdout)
	}
	if !strings.Contains(resp.RunStdout, "content-adhoc-b") {
		t.Fatalf("missing adhoc-b content; stdout=%q", resp.RunStdout)
	}
}
```
