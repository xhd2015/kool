## Expected

- Exit 0.
- Output mentions id `a` and command `echo hi` (props parse succeeded despite spaces).

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
		t.Fatalf("props whitespace should parse; exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab not accepted; out:\n%s", out)
	}
	if !strings.Contains(out, "echo hi") {
		t.Fatalf("missing command echo hi; out:\n%s", out)
	}
	// id=a as token — accept id=a or standalone "a" near plan line
	if !strings.Contains(out, "id=a") && !strings.Contains(out, " a:") &&
		!strings.Contains(out, " a ") && !strings.Contains(out, "- a") &&
		!strings.Contains(out, "id: a") && !strings.Contains(out, "\"a\"") {
		// require at least a clear id marker with "a" and not only inside "echo hi"
		if !strings.Contains(out, " a") && !strings.Contains(out, "a:") {
			t.Fatalf("expected id a in plan; out:\n%s", out)
		}
	}
}
```
