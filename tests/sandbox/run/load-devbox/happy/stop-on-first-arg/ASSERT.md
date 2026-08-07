## Expected

- Sealed run exit 0 without requiring `--` before guest command.
- RunStdout contains `ok-stop-on-first`.
- Load notice present for secondary abs path.

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
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q — StopOnFirstArg should allow guest after --load-devbox",
			resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	if !strings.Contains(resp.RunStdout, "ok-stop-on-first") {
		t.Fatalf("guest echo missing; stdout=%q stderr=%q", resp.RunStdout, resp.RunStderr)
	}
	if len(resp.SecondaryPaths) == 1 {
		notice := "notice: loading devbox " + resp.SecondaryPaths[0]
		if !strings.Contains(resp.RunStdout, notice) {
			t.Fatalf("want load notice %q; stdout=%q", notice, resp.RunStdout)
		}
	}
}
```
