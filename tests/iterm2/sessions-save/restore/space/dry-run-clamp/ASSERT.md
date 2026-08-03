## Expected

- Exit 0
- Plan shows clamped `space 0 (Desktop 1)` (not space 16 / Desktop 17)
- Stderr warning mentions space (clamp / invalid / out of range)
- Not stamped

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "space 0 (Desktop 1)") {
		t.Fatalf("space>=16 must plan space 0 (Desktop 1):\n%s", out)
	}
	// Must not advertise the unclamped desktop.
	if strings.Contains(out, "space 16") || strings.Contains(out, "Desktop 17") {
		t.Fatalf("must not plan unclamped space 16 / Desktop 17:\n%s", out)
	}
	errOut := strings.ToLower(resp.Stderr)
	if !strings.Contains(errOut, "space") {
		t.Fatalf("dry-run clamp should warn on stderr about space; stderr=%q", resp.Stderr)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
