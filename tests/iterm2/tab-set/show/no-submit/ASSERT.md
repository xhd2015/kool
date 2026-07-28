## Expected

- Exit 0.
- Output includes tab `b` command `grok --resume` and a clear `no_submit` marker
  (e.g. `no_submit=true` or equivalent stable token for that tab).
- Tab `a` / `echo a` is listed; output must not claim no_submit for tab a
  (e.g. no blanket "all no_submit"; prefer id=b line carries the flag).

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

	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode != 0 {
		t.Fatalf("show no_submit exit=%d out:\n%s", resp.ExitCode, out)
	}
	if !strings.Contains(out, "grok --resume") {
		t.Fatalf("show missing command for tab b; out:\n%s", out)
	}
	if !strings.Contains(out, "echo a") {
		t.Fatalf("show missing command for tab a; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "no_submit") {
		t.Fatalf("show must surface no_submit for tab b; out:\n%s", out)
	}
	// Prefer explicit true marker.
	if !strings.Contains(lower, "no_submit=true") &&
		!strings.Contains(lower, "no_submit: true") &&
		!strings.Contains(lower, "no_submit true") {
		t.Fatalf("expected no_submit=true (or equivalent) in show output; out:\n%s", out)
	}
}
```
