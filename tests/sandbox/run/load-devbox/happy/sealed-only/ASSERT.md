## Expected

- Build exit 0; secondary built; sealed run exit 0.
- RunStdout contains `notice: loading devbox ` + abs secondary path.
- RunStdout contains `content-from-sealed-load`.

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
	if !strings.Contains(resp.RunStdout, notice) {
		t.Fatalf("want notice %q; stdout=%q", notice, resp.RunStdout)
	}
	if !strings.Contains(resp.RunStdout, "content-from-sealed-load") {
		t.Fatalf("guest should see sealed-load file; stdout=%q", resp.RunStdout)
	}
}
```
