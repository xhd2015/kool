## Expected

- Exit 0
- Stdout contains `space 2 (Desktop 3)`
- Stdout does not show `iterm_window_id` / `4242` as placement
- Would restore present; restored_at still null

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
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	if !strings.Contains(out, "space 2 (Desktop 3)") {
		t.Fatalf("expected space 2 (Desktop 3) in plan:\n%s", out)
	}
	// iterm_window_id is info-only; dry-run must not surface it as placement.
	if strings.Contains(out, "iterm_window_id") {
		t.Fatalf("dry-run must not show iterm_window_id:\n%s", out)
	}
	if strings.Contains(out, "4242") {
		t.Fatalf("dry-run must not show raw iterm_window_id value:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
