## Expected

- Sealed run exit 0.
- RunStdout contains exact prefix `notice: loading devbox ` followed by abs path
  (lowercase `notice:` per contract).
- No ANSI escape sequences (`\x1b`) in RunStdout (doctest capture is non-TTY).

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
	abs := resp.SecondaryPaths[0]
	// Contract: notice: loading devbox <abs> (lowercase notice:)
	want := "notice: loading devbox " + abs
	if !strings.Contains(resp.RunStdout, want) {
		t.Fatalf("RunStdout should contain %q; got %q", want, resp.RunStdout)
	}
	// Prefer notice as its own line (trailing newline after path).
	if !strings.Contains(resp.RunStdout, want+"\n") && !strings.HasSuffix(resp.RunStdout, want) {
		// Soft: still accept if path is present with notice prefix (some tools trim).
		// Hard check already covered by Contains(want).
		_ = want
	}
	if strings.Contains(resp.RunStdout, "\x1b") {
		t.Fatalf("non-TTY capture must not include ANSI; stdout=%q", resp.RunStdout)
	}
	// Reject title-case / wrong label variants only if correct form missing (already checked).
	if strings.Contains(resp.RunStdout, "Notice: loading") && !strings.Contains(resp.RunStdout, "notice: loading") {
		t.Fatalf("notice must use lowercase notice:; got %q", resp.RunStdout)
	}
}
```
