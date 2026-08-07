## Expected

- Build exit 0; secondary pack built; sealed run exit 0.
- RunStdout contains `notice: loading devbox ` + absolute secondary path.
- RunStdout contains secondary file content `content-from-secondary`.
- No ANSI escape sequences in RunStdout (non-TTY capture).

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
		t.Fatalf("build exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if len(resp.SecondaryPaths) != 1 {
		t.Fatalf("expected 1 secondary path; got %v", resp.SecondaryPaths)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	sec := resp.SecondaryPaths[0]
	notice := "notice: loading devbox " + sec
	if !strings.Contains(resp.RunStdout, notice) {
		t.Fatalf("RunStdout should contain %q; got %q", notice, resp.RunStdout)
	}
	if !strings.Contains(resp.RunStdout, "content-from-secondary") {
		t.Fatalf("guest should see secondary file content; stdout=%q", resp.RunStdout)
	}
	if strings.Contains(resp.RunStdout, "\x1b[") || strings.Contains(resp.RunStdout, "\x1b") {
		t.Fatalf("non-TTY capture must not include ANSI escapes; stdout=%q", resp.RunStdout)
	}
}
```
