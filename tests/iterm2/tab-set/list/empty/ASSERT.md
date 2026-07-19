## Expected

- Exit 0.
- Output indicates no sets: empty list, or contains `0` with `set`, or no set names.

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
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := strings.TrimSpace(resp.Stdout + resp.Stderr)
	// Accept empty stdout, or explicit zero-count messaging.
	if out == "" {
		return
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "bots") {
		t.Fatalf("empty dir must not list bots; out:\n%s", out)
	}
	// If any text, prefer zero-set wording
	if strings.Contains(lower, "set") && !strings.Contains(lower, "0") &&
		!strings.Contains(lower, "no ") && !strings.Contains(lower, "none") &&
		!strings.Contains(lower, "empty") {
		// still ok if it's just a header like "tab sets:" with nothing after
		if strings.Contains(lower, "error") {
			t.Fatalf("unexpected error on empty list: %s", out)
		}
	}
}
```
