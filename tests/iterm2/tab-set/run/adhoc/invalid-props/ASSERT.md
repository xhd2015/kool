## Expected

- Non-zero exit.
- Error mentions props / parse / invalid / tab (not silent success).

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
	out := combinedOut(resp)
	if resp.ExitCode == 0 {
		t.Fatalf("expected props parse error; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	// Prefer a parse-related message once implemented; unknown-flag is also RED.
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab not accepted (implementer: wire --tab parse); out:\n%s", out)
	}
	if !strings.Contains(lower, "prop") && !strings.Contains(lower, "parse") &&
		!strings.Contains(lower, "invalid") && !strings.Contains(lower, "tab") &&
		!strings.Contains(lower, "syntax") && !strings.Contains(lower, "malformed") {
		t.Fatalf("expected props/parse error message; out:\n%s", out)
	}
}
```
