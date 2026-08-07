## Expected

- Sealed run exit 0.
- Guest sees load content `content-L-from-load` for `shared.txt`.
- Primary content `content-P-from-primary` must not appear in guest file output
  (may still appear only if mistakenly written; assert absence of primary body
  after notices is via exact content check on cat line).

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
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	if strings.Contains(resp.RunStdout, "content-P-from-primary") {
		t.Fatalf("guest must see load L, not primary P; stdout=%q", resp.RunStdout)
	}
	if !strings.Contains(resp.RunStdout, "content-L-from-load") {
		t.Fatalf("guest should see load content; stdout=%q", resp.RunStdout)
	}
}
```
