## Expected

- Exit 0.
- Plan/output includes default ids `tab-1` and `tab-2` (1-based order).
- Commands `echo one` / `echo two` present.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab not accepted; out:\n%s", out)
	}
	if !strings.Contains(out, "tab-1") {
		t.Fatalf("expected default id tab-1; out:\n%s", out)
	}
	if !strings.Contains(out, "tab-2") {
		t.Fatalf("expected default id tab-2; out:\n%s", out)
	}
	for _, want := range []string{"echo one", "echo two"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q; out:\n%s", want, out)
		}
	}
}
```
