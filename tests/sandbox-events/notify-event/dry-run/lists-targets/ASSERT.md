## Expected

- Exit 0.
- Stdout (or stderr plan output) mentions `sess-dry.sock`.
- Dry-run must not require a live listener (placeholder file is enough).

## Exit Code

- 0

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
		t.Fatalf("dry-run exit=%d want 0; stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	combined := resp.Stdout + "\n" + resp.Stderr
	if !strings.Contains(combined, "sess-dry.sock") {
		t.Fatalf("dry-run should list sess-dry.sock; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
}
```
