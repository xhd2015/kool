## Expected

- Sealed run exit 0.
- Guest sees both PRIM_KEY and LOAD_KEY values as `1/2`.

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
	// Notices may prefix stdout; look for the merged values.
	if !strings.Contains(resp.RunStdout, "1/2") {
		t.Fatalf("want merged env 1/2 in stdout; got %q stderr=%q", resp.RunStdout, resp.RunStderr)
	}
}
```
