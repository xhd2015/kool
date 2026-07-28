## Expected

- Non-zero exit.
- Output mentions not found / unknown / missing / no-such-set.
- Prefer error prefix style `tab-set update:` (soft).

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

	// Classic TDD: fail while update is not a known subcommand (do not
	// treat "unknown subcommand" as the scenario-specific error).
	if strings.Contains(strings.ToLower(out), "unknown subcommand") {
		t.Fatalf("update not implemented yet (RED until implementer); out:\n%s", out)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("missing set must fail; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized argument") {
		t.Fatalf("update not routed (open-dir fallback); out:\n%s", out)
	}
	if !strings.Contains(lower, "no-such-set") &&
		!strings.Contains(lower, "not found") &&
		!strings.Contains(lower, "unknown") &&
		!strings.Contains(lower, "missing") &&
		!strings.Contains(lower, "no such") {
		t.Fatalf("expected missing-set error; out:\n%s", out)
	}
}
```
