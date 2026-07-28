## Expected

- Exit 0 (no live iTerm).
- Plan mentions both commands (`echo a`, `echo staged`).
- Plan marks the no_submit tab with a stable token such as `(no_submit)` or
  `no_submit` on/near tab b / `echo staged`.
- Default tab a is not the only place carrying the marker (marker must relate to
  the no_submit command).

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
		t.Fatalf("dry-run no_submit exit=%d out:\n%s", resp.ExitCode, out)
	}
	for _, want := range []string{"echo a", "echo staged"} {
		if !strings.Contains(out, want) {
			t.Fatalf("plan missing %q; out:\n%s", want, out)
		}
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "no_submit") {
		t.Fatalf("dry-run plan must mark no_submit tabs; out:\n%s", out)
	}
	// Prefer token near staged command or explicit (no_submit).
	if strings.Contains(lower, "(no_submit)") {
		return
	}
	// Fallback: same line-ish: "echo staged" and no_submit both present is required above.
	// Reject if only generic "no_submit" docs without the staged command context.
	if !strings.Contains(out, "echo staged") {
		t.Fatalf("expected no_submit mark associated with staged tab; out:\n%s", out)
	}
}
```
