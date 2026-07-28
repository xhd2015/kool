## Expected

- Non-zero exit.
- Error mentions `no_submit` and rejects the **value** `maybe` (not silent success).
- Must not fail only as unknown prop key — once `no_submit` is a known key,
  message should indicate invalid bool/value (true/false/1/0/yes/no).

## Exit Code

- ≠ 0

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected invalid no_submit value error; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab not accepted; out:\n%s", out)
	}
	if !strings.Contains(lower, "no_submit") {
		t.Fatalf("error should mention no_submit; out:\n%s", out)
	}
	// Reject "unknown key" only — that means the key is not wired yet (RED).
	if strings.Contains(lower, "unknown key") {
		t.Fatalf("no_submit must be a known prop; want invalid-value error for maybe; out:\n%s", out)
	}
	// Value-oriented error once key is accepted.
	if !strings.Contains(lower, "maybe") && !strings.Contains(lower, "invalid") &&
		!strings.Contains(lower, "bool") && !strings.Contains(lower, "true") &&
		!strings.Contains(lower, "false") {
		t.Fatalf("expected invalid-value style message for no_submit=maybe; out:\n%s", out)
	}
}
```