## Expected

- Non-zero exit.
- Stderr/stdout mentions error / not found / missing set name.

## Exit Code

- ≠ 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for missing set; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(resp.Stderr + resp.Stdout)
	// Reject open-dir fallback ("unrecognized arguments") — tab-set must handle show.
	if strings.Contains(out, "unrecognized argument") || strings.Contains(out, "unrecognized flag") {
		t.Fatalf("tab-set show not routed (open-dir fallback); out:\n%s", resp.Stderr+resp.Stdout)
	}
	if !strings.Contains(out, "no-such-set") &&
		!strings.Contains(out, "not found") &&
		!strings.Contains(out, "unknown") &&
		!strings.Contains(out, "missing") {
		t.Fatalf("expected missing-set error message; out:\n%s", resp.Stderr+resp.Stdout)
	}
}
```

