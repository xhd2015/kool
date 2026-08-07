## Expected

- Sealed run exit 0; load file visible.
- Exactly one `notice: loading devbox <abs>` for the shared path (first-seen dedupe).

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
	if len(resp.SecondaryPaths) != 1 {
		t.Fatalf("expected 1 secondary; got %v", resp.SecondaryPaths)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	notice := "notice: loading devbox " + resp.SecondaryPaths[0]
	n := strings.Count(resp.RunStdout, notice)
	if n != 1 {
		t.Fatalf("want exactly 1 notice for deduped path; count=%d notice=%q stdout=%q", n, notice, resp.RunStdout)
	}
	if !strings.Contains(resp.RunStdout, "content-dedupe-load") {
		t.Fatalf("guest should see load file; stdout=%q", resp.RunStdout)
	}
}
```
