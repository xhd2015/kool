## Expected

- Sealed run exit 0.
- Guest sees `content-from-nested-a` (from nested sealed pack A via B).
- Notices appear for loaded boxes (at least B and A abs paths).

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
		t.Fatalf("expected secondaries A,B; got %v", resp.SecondaryPaths)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	if !strings.Contains(resp.RunStdout, "content-from-nested-a") {
		t.Fatalf("guest should see nested A file; stdout=%q", resp.RunStdout)
	}
	absA := resp.SecondaryPaths[0]
	absB := resp.SecondaryPaths[1]
	// Both nested load targets should be announced on successful load.
	if !strings.Contains(resp.RunStdout, "notice: loading devbox "+absB) {
		t.Fatalf("want notice for B %q; stdout=%q", absB, resp.RunStdout)
	}
	if !strings.Contains(resp.RunStdout, "notice: loading devbox "+absA) {
		t.Fatalf("want notice for nested A %q; stdout=%q", absA, resp.RunStdout)
	}
}
```
